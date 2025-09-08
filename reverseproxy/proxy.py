#!/usr/bin/env python3
import asyncio
import aiohttp
from aiohttp import web
import sys
import json
import os
import traceback

CHUNK_SIZE = 1024 * 1024  # Default chunk size: 1MB
m3u = "http://95.142.46.84:5665/playlistall/all.m3u"
# http://0.0.0.0:8080/stream/Wednesday.S02.1080p.NF.WEB-DL-EniaHD.m3u?link=e143d60dabde0af9d263abab362e870201fc8acf&m3u&fn=file.m3u
# http://0.0.0.0:8080/stream/Wednesday.S02E01.Here.We.Woe.Again.1080p.NF.WEB-DL-EniaHD.mkv?link=e143d60dabde0af9d263abab362e870201fc8acf&index=1&play
# Response cache: {(method, path, query, body): (status, headers, body)}
lock = asyncio.Lock()

class CacheEntry:
    def __init__(self, headReponseHeaders, url, headers, params, method):
        self.initialResponseHeaders = headReponseHeaders
        self.chunks: dict[int, bytes] = {}
        self.len = int(headReponseHeaders['Content-Length'])
        self.url = url
        self.headers = headers
        self.params = params
        self.method = method

    def set(self, offset, chunk):
        # TODO Verify overlapping chinks consistency
        # TODO Merge
        self.chunks[offset] = chunk
    def get(self, offset):
        async def get_chunk_length(offset):
            while True:
                if offset >= self.len:
                    break
                found_key = max((k for k in self.chunks.keys() if k <= offset), default=None)

                if found_key is None:
                    print(f"Waiting {offset}")
                    await asyncio.sleep(1)
                    continue
                yield self.chunks[found_key]
                offset += len(self.chunks[found_key])

        return get_chunk_length(offset)

class Cache:
    def __init__(self):
        self.store = {}
    async def getOrCreate(self, key, headResponseHeaders, method, url, headers, params) -> CacheEntry:
        async with lock:
            if key not in self.store:
                self.store[key] = CacheEntry(headReponseHeaders=headResponseHeaders, method=method, url=url, headers=headers, params=params)
        return self.store.get(key)

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
            response = web.StreamResponse(status=resp.status, headers=resp.headers)
            await response.prepare(request)
            async for chunk in resp.content.iter_chunked(CHUNK_SIZE):
                await response.write(chunk)
            await response.write_eof()
            return response

def find_hole(chunks, size):
    if len(chunks) == 0:
        return (0, size)
    sorted_chunks = sorted(chunks.keys())
    for i in range(len(sorted_chunks) - 1):
        if sorted_chunks[i] + len(chunks[sorted_chunks[i]]) < sorted_chunks[i + 1]:
            return (sorted_chunks[i] + len(chunks[sorted_chunks[i]]), sorted_chunks[i + 1])
    if sorted_chunks[-1] + len(chunks[sorted_chunks[-1]]) < size:
        return (sorted_chunks[-1] + len(chunks[sorted_chunks[-1]]), size)
    return (None, None)

async def downloader():
    print("Downloader started")
    while True:
        await asyncio.sleep(1)
        async with lock:
            keys = response_cache.store.keys()
            for key in keys:
                cached: CacheEntry = response_cache.store[key]
                (startOffset, endOffset) = find_hole(cached.chunks, cached.len)
                if startOffset is None:
                    continue
                endOffset = min(endOffset, startOffset + CHUNK_SIZE * 10)
                # while len(cached.chunks) > 1:
                #     first = sorted(cached.chunks.keys())[0]
                #     second = sorted(cached.chunks.keys())[1]
                #     if first + len(cached.chunks[first]) == second:
                #         print(f"Merging chunks {first} and {second} for key {key}")
                #         cached.chunks[first] += cached.chunks[second]
                #         del cached.chunks[second]
                #     else:
                #         break
                # if len(cached.chunks) == 0:
                #     startOffset = 0
                #     endOffset = cached.len
                # else:
                #     startKey = sorted(cached.chunks.keys())[0]
                #     startOffset = startKey + cached.chunks[startKey]
                #     if len(cached.chunks) > 1:
                #         startKey = sorted(cached.chunks.keys())[1]
                #         endOffset = startKey + cached.chunks[startKey]
                #     else:
                #         endOffset = cached.len
                print(f"Downloading startOffset {startOffset}..{endOffset} for key {key}")
                rangeHeader = {"Range": f"bytes={startOffset}-{endOffset-1}"}
                async with aiohttp.ClientSession() as session:
                    # TODO data=body
                    headers_with_range = dict(cached.headers)
                    headers_with_range.update(rangeHeader)
                    async with session.request(cached.method, cached.url, headers=headers_with_range, params=cached.params) as resp:

                        # Stream response in chunks
                        print(f"Response headers: {resp.headers}")
                        if "Content-Range" not in resp.headers:
                            raise ValueError("No Content-Range in response for Range request")
                        if "bytes" not in resp.headers["Content-Range"]:
                            raise ValueError("Only bytes Content-Range supported")
                        (start_end, size) = resp.headers["Content-Range"].replace("bytes ", "").split("/")
                        if start_end == "*":
                            (start, end) = (0, size - 1)
                        else:
                            (start, end) = [int(i) for i in start_end.split("-")]
                        print(f"Content-Range: {start}-{end}/{size}")
                        offset = start
                        async for chunk in resp.content.iter_chunked(CHUNK_SIZE):
                            cached.set(offset, chunk)
                            offset += len(chunk)


async def proxy_handler(request):
    try:
        backend_url = request.app['backend_url']
        headers = dict(request.headers)
        params = request.rel_url.query
        path = request.rel_url.path
        method = request.method
        body = await request.read()
        cache_key = (method, path, tuple(sorted(params.items())), body)
        if "Range" not in headers.keys():
            print(f"Simple handler {cache_key}")
            return await simple_proxy_handler(request)
        range = headers.get("Range")
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
            print(f"HEAD response status: {head_resp.status}")
            print(f"HEAD response headers: {head_resp.headers}")
            headHeaders = head_resp.headers
        cache_key = (method, path, tuple(sorted(params.items())), body, tuple(sorted(headHeaders.items())))                
        print(f"Cache key: {cache_key}")
        print(f"Range requested: {range}")
        cached = await response_cache.getOrCreate(cache_key, headHeaders, method=method, url = backend_url + path, headers=head_headers, params=params)
        stream = cached.get(range[0] or 0)

        response = web.StreamResponse(status=206, headers=headHeaders)
        await response.prepare(request)
        async for chunk in stream:
            await response.write(chunk)
        await response.write_eof()
        return response

    except aiohttp.ClientConnectionResetError as e:
        print(f"ClientConnectionResetError: {e}", file=sys.stderr)
        return web.Response(status=500, text=f"Internal Server Error: {e}")
    except Exception as e:
        print(f"proxy_handler error: {e}", file=sys.stderr)
        traceback.print_exc()
        return web.Response(status=500, text=f"Internal Server Error: {e}")


async def server(app):
    runner = web.AppRunner(app)  
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()   


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
    loop.create_task(server(app))

    loop.run_forever()

    
if __name__ == "__main__":
    main()
