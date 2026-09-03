import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const mime = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2',
  '.webmanifest': 'application/manifest+json',
  '.xml': 'application/xml',
}

/** Serve the embedded SPA so Playwright can mock same-origin APIs. */
export function serveEmbedDir(rootDir) {
  const root = path.resolve(rootDir)
  const server = http.createServer((req, res) => {
    let urlPath = '/'
    try {
      urlPath = decodeURIComponent(new URL(req.url || '/', 'http://127.0.0.1').pathname)
    } catch {
      res.writeHead(400)
      res.end()
      return
    }
    if (urlPath === '/') urlPath = '/index.html'
    const file = path.normalize(path.join(root, urlPath))
    if (!file.startsWith(root)) {
      res.writeHead(403)
      res.end()
      return
    }
    fs.readFile(file, (err, data) => {
      if (err) {
        res.writeHead(404)
        res.end()
        return
      }
      res.writeHead(200, { 'content-type': mime[path.extname(file)] || 'application/octet-stream' })
      res.end(data)
    })
  })
  return new Promise((resolve, reject) => {
    server.listen(0, '127.0.0.1', () => {
      const addr = server.address()
      if (!addr || typeof addr === 'string') {
        reject(new Error('failed to bind embed static server'))
        return
      }
      resolve({ server, origin: `http://127.0.0.1:${addr.port}` })
    })
    server.on('error', reject)
  })
}

export function embedPagesRoot() {
  return path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../server/web/pages/template/pages')
}
