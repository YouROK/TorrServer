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


# class CacheEntry:

class Cache:
    def __init__(self):
        self.store = {}
    def get(self, key): # -> CacheEntry | None:
        return self.store.get(key)
    def set(self, key, offset, value):
        self.store[key] = value

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
        range = [[to_int_or_none(i) for i in r.split("-")] for r in range.split(",")]
        print(f"Cache key: {cache_key}")
        print(f"Range requested: {range}")
        # cached = response_cache.get(cache_key)
        # if cached:
        #     status, resp_headers, resp_body = cached
        #     print(f"Cache hit for {method} {path}")
        #     return web.Response(status=status, headers=resp_headers, body=resp_body)

        async with aiohttp.ClientSession() as session:
            async with session.request(method, backend_url + path, headers=headers, params=params, data=body) as resp:

                # Stream response in chunks
                response = web.StreamResponse(status=resp.status, headers=resp.headers)
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
                await response.prepare(request)
                offset = start
                async for chunk in resp.content.iter_chunked(CHUNK_SIZE):
                    response_cache.set(cache_key, offset, chunk)
                    await response.write(chunk)
                    offset += len(chunk)
                await response.write_eof()
                return response
    except aiohttp.ClientConnectionResetError as e:
        print(f"ClientConnectionResetError: {e}", file=sys.stderr)
        return web.Response(status=500, text=f"Internal Server Error: {e}")
    except Exception as e:
        print(f"proxy_handler error: {e}", file=sys.stderr)
        traceback.print_exc()
        return web.Response(status=500, text=f"Internal Server Error: {e}")

def main():
    if len(sys.argv) < 2:
        backend_url = "http://95.142.46.84:5665"
    else:
        backend_url = sys.argv[1]
    app = web.Application()
    app['backend_url'] = backend_url
    app.router.add_route('*', '/{tail:.*}', proxy_handler)
    web.run_app(app, port=8080)

if __name__ == "__main__":
    main()
