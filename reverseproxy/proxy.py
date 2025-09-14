#!/usr/bin/env python3
import asyncio
from tracemalloc import start
import aiohttp
from aiohttp import web
import sys
import json
import os
import traceback
from sortedcontainers import SortedDict

from throttler import DownloadSpeedLimiter

CHUNK_SIZE = 1024 * 1024  # Default chunk size: 1MB


cache_path = "cache"
if not os.path.exists(cache_path):
    os.makedirs(cache_path)
class Chunk:
    def __init__(self, path, offset, data = None):
        self.path = path
        self.offset = offset
        self.file = os.path.join(cache_path, path, str(offset).zfill(15))
        if data is not None:
            with open(self.file, "wb") as f:
                f.write(data)
        self._len = os.path.getsize(self.file)
    def len(self):
        return self._len

    def data(self, start = 0, size = -1):
        with open(self.file, "rb") as f:
            f.seek(start)
            dt = f.read(min(self.len() - start, size) if size >= 0 else None)
            return dt
    def append(self, data):
        with open(self.file, "r+b") as f:
            f.seek(0, os.SEEK_END)
            f.write(data)
            f.flush()
        self._len = os.path.getsize(self.file)
    
    def __repr__(self):
        return f"Chunk(offset={self.offset}, file={self.file}, length={self.len()})"


priority = {}

class CacheEntry:
    def __init__(self, key, headReponseHeaders, url, headers, params, method):
        self.lock = asyncio.Lock()
        self.key = key
        self.initialResponseHeaders = headReponseHeaders
        
        self.len = int(headReponseHeaders['Content-Length'])
        self.url = url
        self.headers = headers
        self.params = params
        self.method = method
        self.cachePath = (self.url + self.initialResponseHeaders['Etag'][:20]).replace("/", "_").replace(":", "_").replace("?", "_").replace("&", "_").replace("=", "_")
        os.makedirs(os.path.join(cache_path, self.cachePath), exist_ok=True)
        self.chunks: SortedDict[int, Chunk] = SortedDict({
            int(l): Chunk(self.cachePath, int(l), None) 
            for l in os.listdir(os.path.join(cache_path, self.cachePath))
        })
        # print(f"CacheEntry created for {url} with {self.chunks} chunks")

    async def set(self, offset, chunk: bytes):
        async with self.lock:
            self.chunks[offset] = Chunk(self.cachePath, offset, chunk)
            return self.chunks[offset]

    def get(self, offset):
        async def get_chunk_length(offset: int):
            while True:
                await asyncio.sleep(0)
                if offset >= self.len:
                    break
                async with self.lock:
                    found_index = self.chunks.bisect_right(offset) - 1
                    # print(f"Found index {found_index} for offset {offset}")
                    found_chunk: Chunk = None
                    if found_index >= 0:
                        found_chunk = self.chunks[self.chunks.keys()[found_index]]
                        # print(f"Found chunk {found_chunk}, {found_chunk.offset}, {found_chunk.len()}")
                    if found_chunk is not None and found_chunk.offset + found_chunk.len() <= offset:
                        found_chunk = None
                        
                    # found_key = max((k for k, c in self.chunks.items() if k <= offset and c.offset + c.len() > offset), default=None)
                # print(f"Raising priority {self.key}, {offset}")
                if found_chunk is None:
                    # print(f"Waiting {offset}")
                    # log = {k: (c.offset, c.offset + c.len()) for k, c in self.chunks.items()}
                    # print(f"Chunk keys {log}")
                    await asyncio.sleep(1)
                    yield b""
                    continue
    
                dt = found_chunk.data(offset - found_chunk.offset, CHUNK_SIZE)
                if len(dt) == 0:
                    raise ValueError("Empty chunk data")
                yield dt
                offset += len(dt)

        return get_chunk_length(offset)

class Cache:
    def __init__(self):
        self.store: dict[any, CacheEntry] = {}
        self.lock = asyncio.Lock()

    async def allItems(self) -> dict[any, CacheEntry]:
        async with self.lock:
            return dict(self.store)
        
    async def getOrCreate(self, key, headResponseHeaders, method, url, headers, params) -> CacheEntry:
        async with self.lock:
            if key not in self.store:
                self.store[key] = CacheEntry(key, headReponseHeaders=headResponseHeaders, method=method, url=url, headers=headers, params=params)
            return self.store.get(key)
    async def mergeAnyTwo(self):
        async with self.lock:
            for entry in self.store.values():
                async with entry.lock:
                    keysI = list(entry.chunks.keys())
                    for keyI in range(len(keysI) - 1):
                        chunk1:Chunk = entry.chunks[keysI[keyI]]
                        chunk2:Chunk = entry.chunks[keysI[keyI+1]]
                        if chunk2 is None:
                            print(f"No chunk2 for {keysI[keyI]} in {entry.key}")
                            continue
                        if chunk1.offset >= chunk2.offset:
                            print(f"Chunk1 {chunk1} is not before Chunk2 {chunk2} in {entry.key}")
                            continue
                        if chunk1.offset + chunk1.len() - chunk2.offset >= chunk2.len():
                            print(f"Chunk1 {chunk1} already covers Chunk2 {chunk2} in {entry.key}: {chunk1.offset + chunk1.len() - chunk2.offset >= chunk2.len()}")
                            continue
                        if chunk1.offset + chunk1.len() < chunk2.offset:
                            print(f"Chunk1 {chunk1} has gap after Chunk2 {chunk2} in {entry.key}: {chunk1.offset + chunk1.len() - chunk2.offset >= chunk2.len()}")
                            continue
                        print(f"Merging {chunk1}+{chunk2}[{chunk1.offset + chunk1.len() - chunk2.offset}:] from {entry.key}")
                        chunk1.append(chunk2.data(chunk1.offset + chunk1.len() - chunk2.offset, -1))
                        del entry.chunks[chunk2.offset]
                        os.remove(chunk2.file)
                        return True
        return False

response_cache = Cache()

async def simple_proxy_handler(request):
    backend_url = request.app['backend_url']
    headers = dict(request.headers)
    params = request.rel_url.query
    path = request.rel_url.path
    method = request.method
    body = await request.read()

    async with aiohttp.ClientSession() as session:
        async with session.request(method, backend_url + path, headers=headers, params=params, data=body) as resp:
            # print("Response status:", resp.headers)
            response = web.StreamResponse(status=resp.status, headers=resp.headers)
            await response.prepare(request)
            async for chunk in resp.content.iter_chunked(CHUNK_SIZE):
                await response.write(chunk)
            await response.write_eof()
            return response

def find_hole(entry: CacheEntry, start_offset=0):
    chunks = entry.chunks
    if len(chunks) == 0:
        return (0, entry.len)
    # TODO Lock with merger
    keys = chunks.bisect_right(start_offset) - 1
    keys = list(chunks.islice(keys, len(chunks)))
    for i in range(len(keys) - 1):
        if keys[i + 1] < start_offset:
            continue
        curr: Chunk = chunks[keys[i]]
        if curr.offset + curr.len() < keys[i + 1]:
            return (curr.offset + curr.len(), keys[i + 1])
    if keys[-1] + chunks[keys[-1]].len() < entry.len:
        return (keys[-1] + chunks[keys[-1]].len(), entry.len)
    return (None, None)

async def merger():
    while True:
        try:
            cnt = 0
            if await response_cache.mergeAnyTwo() and cnt < 1000:
                cnt += 1
            if cnt > 0:
                await asyncio.sleep(0.1)
                continue
        except Exception as e:
            print(f"merger error: {e}", file=sys.stderr)
            traceback.print_exc()
        await asyncio.sleep(30)

async def verifirer():
    # TODO
    # TODO Locks
    pass

async def downloader():
    global priority
    limiter = DownloadSpeedLimiter(200000)
    while True:
        try:
            async with aiohttp.ClientSession() as session:
                while True:
                    print(f"C1")
                    await asyncio.sleep(0)
                    keys = await response_cache.allItems()
                    queuekeys = [(k, v, priority[k]) for k, v in keys.items() if k in priority.keys() and find_hole(v, priority[k])[0] is not None]
                    print(f"C1.1")
                    queuekeys.sort(key=lambda x: find_hole(x[1], x[2])[0] - x[2])
                    print(f"C1.2")
                    queuekeys += [(k, v, None) for k, v in keys.items()]
                    print(f"C1.3")
                    queuekeys = [(k,v,p) for k,v,p in queuekeys if find_hole(v, p or 0)[0] is not None]
                    
                    if len(queuekeys) == 0:
                        print(f"Nothing to download, sleeping")
                        await asyncio.sleep(10)
                        continue
                    print(f"C2")
                    priority = {}
                    for key, cached, startFrom in queuekeys:
                        startOffset = None
                        endOffset = None
                        if startFrom is not None:
                            (startOffset, endOffset) = find_hole(cached, startFrom)
                            if startOffset is not None:
                                startOffset = max(startOffset, startFrom)
                                endOffset = startOffset + CHUNK_SIZE * 10
                        print(f"C3")

                        if startOffset is None:
                            (startOffset, endOffset) = find_hole(cached)
                        print(f"C4")

                        if startOffset is None:
                            continue
                        endOffset = min(endOffset, startOffset + CHUNK_SIZE)

                        print(f"Downloading startOffset {startOffset}..{endOffset} for key {key} ({startOffset * 100.0/cached.len}%)")
                        rangeHeader = {"Range": f"bytes={startOffset}-{endOffset-1}"}
                        headers_with_range = dict(cached.headers)
                        headers_with_range.update(rangeHeader)
                        async with session.request(cached.method, cached.url, headers=headers_with_range, params=cached.params) as resp:

                            # Stream response in chunks
                            # print(f"Response headers: {resp.headers}")
                            if "Content-Range" not in resp.headers:
                                raise ValueError("No Content-Range in response for Range request")
                            if "bytes" not in resp.headers["Content-Range"]:
                                raise ValueError("Only bytes Content-Range supported")
                            (start_end, size) = resp.headers["Content-Range"].replace("bytes ", "").split("/")
                            if start_end == "*":
                                (start, end) = (0, size - 1)
                            else:
                                (start, end) = [int(i) for i in start_end.split("-")]
                            # print(f"Content-Range: {start}-{end}/{size}")
                            # 1newChunk: Chunk = None
                            receivedBytes: bytes = b""
                            receivedBytesTotal = 0

                            async for receivedBytesLoc in resp.content.iter_chunked(CHUNK_SIZE):
                                receivedBytes += receivedBytesLoc
                                receivedBytesTotal += len(receivedBytesLoc)

                            await cached.set(start, receivedBytes) # 2
                            limiter.consumed(receivedBytesTotal)
                            delay = limiter.delay()
                            print(f"Sleep {delay}, {receivedBytesTotal}, est {limiter._estimated_speed} bytes/second")
                            if startFrom is None:
                                await asyncio.sleep(delay)
                            else:
                                print(f"Skip sleep for priority download")
                                await asyncio.sleep(0)
                            
                        print(f"Downloading done")
                        break
        except Exception as e:
            print(f"downloader error: {e}", file=sys.stderr)
            traceback.print_exc()

async def proxy_handler(request):
    try:
        backend_url = request.app['backend_url']
        headers = dict(request.headers)
        params = request.rel_url.query
        path = request.rel_url.path
        method = request.method
        body = await request.read()
        # if "Range" not in headers.keys():
        if "stream" not in path or "m3u" in path:
            # print(f"Simple handler {method}, {path}, {tuple(sorted(params.items()))}, {body}, {headers}")
            try:
                return await simple_proxy_handler(request)
            except Exception as e:
                print(f"simple_proxy_handler error: {e}", file=sys.stderr)
                # traceback.print_exc()
                return web.Response(status=500, text=f"Internal Server Error: {e}")
        
        if "Range" not in headers.keys():
            headers["Range"] = "bytes=0-"
            
        range = headers.get("Range")
        if range is not None:
            if "bytes" not in range:
                raise ValueError("Only bytes range supported")
            range = range.replace("bytes=", "")
            def to_int_or_none(s):
                try:
                    return int(s)
                except ValueError:
                    return None
            range = [to_int_or_none(i) for r in range.split(",") for i in r.split("-") ]
        
        # Make a HEAD request to get resource info
        async with aiohttp.ClientSession() as session:
            head_headers = dict(headers)
            head_headers.pop("Range", None)
            head_resp = await session.head(backend_url + path, headers=head_headers, params=params)
            # print(f"HEAD response status: {head_resp.status}")
            # print(f"HEAD response headers: {head_resp.headers}")
            headHeaders = head_resp.headers
            if head_resp.status != 200:
                return web.Response(status=head_resp.status, text=f"Upstream server returned status {head_resp.status}")

        cache_key = (method, path, (headHeaders['Etag'][:20].__hash__()))
        # cache_key = (method, path, (tuple(sorted(params.items())), body, tuple(sorted(headHeaders.items()))).__hash__())
        print(f"Cache key: {cache_key}")
        # print(f"Range requested: {range}")
        cached = await response_cache.getOrCreate(cache_key, headHeaders, method=method, url = backend_url + path, headers=head_headers, params=params)

        start = (range and range[0]) or 0
        headHeaders = dict(headHeaders)
        if range is not None:
            end = range[1] if len(range) > 1 and range[1] is not None else cached.len - 1
            content_range_header = f"bytes {start}-{end}/{cached.len}"
            headHeaders["Content-Range"] = content_range_header
            headHeaders["Content-Length"] = f"{end - start + 1}"
            headHeaders["Accept-Ranges"] = 'bytes'
        # print(f"Serving with headers: {headHeaders}")

        response = web.StreamResponse(status=206, headers=headHeaders)
        await response.prepare(request)
        # print("Sending chunks")
        stream = cached.get(start)
        leftBytes = (end - start + 1) if range is not None else cached.len - start
        async for chunk in stream:
            if 'NoPriority' not in headers.keys():
                priority[cache_key] = start
            await response.write(chunk[:min(len(chunk), leftBytes)])
            leftBytes -= len(chunk)
            # print(f"Sending chunks {cache_key} {len(chunk)}")
            if leftBytes <= 0:
                break
            await asyncio.sleep(0)
        await response.write_eof()
        return response

    except aiohttp.ClientConnectionResetError as e:
        print(f"ClientConnectionResetError: {e}", file=sys.stderr)
        return web.Response(status=500, text=f"Internal Server Error: {e}")
    # except aiohttp.ConnectionResetError as e:
    #     print(f"ConnectionResetError: {e}", file=sys.stderr)
    #     return web.Response(status=500, text=f"Internal Server Error: {e}")
    except Exception as e:
        print(f"proxy_handler error: {e}", file=sys.stderr)
        traceback.print_exc()
        return web.Response(status=500, text=f"Internal Server Error: {e}")


async def server(app):
    runner = web.AppRunner(app)  
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()   

async def root_m3u_downloader(m3u_url):
    from m3u_download import process_m3u
    await asyncio.sleep(2)
    while True:
        await process_m3u(m3u_url, None)
        print(f"Root M3U downloader sleeping for 1 hour")
        await asyncio.sleep(10 * 60)

def main():
    if len(sys.argv) < 2:
        backend_url = "http://95.142.46.84:5665"
    else:
        backend_url = sys.argv[1]
    app = web.Application()
    app['backend_url'] = backend_url
    app.router.add_route('*', '/{tail:.*}', proxy_handler)
    loop = asyncio.new_event_loop()
    loop.create_task(downloader())
    loop.create_task(merger())
    loop.create_task(server(app))
# http://0.0.0.0:8080/stream/Wednesday.S02.1080p.NF.WEB-DL-EniaHD.m3u?link=e143d60dabde0af9d263abab362e870201fc8acf&m3u&fn=file.m3u
# http://0.0.0.0:8080/playlistall/all.m3u
# http://0.0.0.0:8080/stream/Wednesday.S02.1080p.NF.WEB-DL-EniaHD.m3u?link=e143d60dabde0af9d263abab362e870201fc8acf&m3u&fn=file.m3u
# http://0.0.0.0:8080/stream/Wednesday.S02E01.Here.We.Woe.Again.1080p.NF.WEB-DL-EniaHD.mkv?link=e143d60dabde0af9d263abab362e870201fc8acf&index=1&play

    loop.create_task(root_m3u_downloader("http://0.0.0.0:8080/playlistall/all.m3u"))
    # loop.create_task(root_m3u_downloader("http://0.0.0.0:8080/stream/Tears%20of%20Steel.m3u?link=209c8226b299b308beaf2b9cd3fb49212dbd13ec&m3u"))
    loop.run_forever()

    
if __name__ == "__main__":
    main()
