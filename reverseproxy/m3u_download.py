#!/usr/bin/env python3

import os
import sys
import asyncio
import aiohttp
import aiofiles
import urllib.parse
from urllib.parse import urljoin, urlparse
from tqdm.asyncio import tqdm


def is_url(string):
    """Check if the provided string is a URL."""
    try:
        result = urlparse(string)
        return all([result.scheme, result.netloc])
    except:
        return False


async def download_file(session, url, destination):
    """Download a file from URL to the specified destination asynchronously."""
    try:
        os.makedirs(os.path.dirname(destination), exist_ok=True)
        if os.path.exists(destination):
            print(f"File already exists: {destination}")
            return True
        print(f"Downloading: {url} -> {destination}")
        async with session.get(url) as response:
            response.raise_for_status()
            total_size = int(response.headers.get('content-length', 0))
            block_size = 8192
            async with aiofiles.open(destination, 'wb') as f:
                pbar = tqdm(total=total_size, unit='B', unit_scale=True, desc=os.path.basename(destination))
                async for chunk in response.content.iter_chunked(block_size):
                    if chunk:
                        await f.write(chunk)
                        pbar.update(len(chunk))
                pbar.close()
        return True
    except Exception as e:
        print(f"Error downloading {url}: {str(e)}")
        return False


async def process_m3u(m3u_url, output_dir):
    """Process an M3U file and download all the files asynchronously."""
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(m3u_url) as response:
                response.raise_for_status()
                content = await response.text()
            base_url = m3u_url.rsplit('/', 1)[0] + '/'
            tasks = []
            for line in content.splitlines():
                line = line.strip()
                if not line or line.startswith('#'):
                    continue
                file_url = line if is_url(line) else urljoin(base_url, line)
                path_part = urlparse(file_url).path
                if path_part.startswith('/'):
                    path_part = path_part[1:]
                path_part = urllib.parse.unquote(path_part)
                local_path = os.path.join(output_dir, path_part)
                tasks.append(download_file(session, file_url, local_path))
            await asyncio.gather(*tasks)
        return True
    except Exception as e:
        print(f"Error processing M3U file: {str(e)}")
        return False


def main():
    if len(sys.argv) < 2:
        print("Usage: python m3u_download.py <m3u_url> [output_directory]")
        sys.exit(1)
    m3u_url = sys.argv[1]
    output_dir = sys.argv[2] if len(sys.argv) > 2 else '.'
    print(f"Downloading content from {m3u_url} to {output_dir}")
    asyncio.run(process_m3u(m3u_url, output_dir))


if __name__ == "__main__":
    main()