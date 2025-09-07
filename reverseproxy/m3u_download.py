#!/usr/bin/env python3

import os
import sys
import requests
import urllib.parse
from urllib.parse import urljoin, urlparse
from tqdm import tqdm

def is_url(string):
    """Check if the provided string is a URL."""
    try:
        result = urlparse(string)
        return all([result.scheme, result.netloc])
    except:
        return False

def download_file(url, destination):
    """Download a file from URL to the specified destination."""
    try:
        # Create directories if they don't exist
        os.makedirs(os.path.dirname(destination), exist_ok=True)
        
        # Don't download if file already exists
        if os.path.exists(destination):
            print(f"File already exists: {destination}")
            return True
        
        # Download the file
        print(f"Downloading: {url} -> {destination}")
        response = requests.get(url, stream=True)
        response.raise_for_status()
        
        # Get file size for progress bar
        total_size = int(response.headers.get('content-length', 0))
        block_size = 8192
        
        with open(destination, 'wb') as f:
            with tqdm(total=total_size, unit='B', unit_scale=True, desc=os.path.basename(destination)) as pbar:
                for chunk in response.iter_content(chunk_size=block_size):
                    if chunk:
                        f.write(chunk)
                        pbar.update(len(chunk))
        
        return True
    except Exception as e:
        print(f"Error downloading {url}: {str(e)}")
        return False

def process_m3u(m3u_url, output_dir):
    """Process an M3U file and download all the files."""
    try:
        # Download the M3U file
        response = requests.get(m3u_url)
        response.raise_for_status()
        content = response.text
        
        # Parse the base URL for relative paths
        base_url = m3u_url.rsplit('/', 1)[0] + '/'
        
        # Process each line
        for line in content.splitlines():
            line = line.strip()
            
            # Skip empty lines and comments
            if not line or line.startswith('#'):
                continue
            
            # Create full URL if it's a relative path
            file_url = line if is_url(line) else urljoin(base_url, line)
            
            # Extract the path part from the URL
            path_part = urlparse(file_url).path
            if path_part.startswith('/'):
                path_part = path_part[1:]
            
            # URL decode the filename to handle special characters
            path_part = urllib.parse.unquote(path_part)
            
            # Create local destination path
            local_path = os.path.join(output_dir, path_part)
            
            # Download the file
            download_file(file_url, local_path)
    
    except Exception as e:
        print(f"Error processing M3U file: {str(e)}")
        return False
    
    return True

def main():
    if len(sys.argv) < 2:
        print("Usage: python download.py <m3u_url> [output_directory]")
        sys.exit(1)
    
    m3u_url = sys.argv[1]
    output_dir = sys.argv[2] if len(sys.argv) > 2 else '.'
    
    print(f"Downloading content from {m3u_url} to {output_dir}")
    process_m3u(m3u_url, output_dir)

if __name__ == "__main__":
    main()