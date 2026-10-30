import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const read = (relative: string) => readFileSync(fileURLToPath(new URL(relative, import.meta.url)), 'utf8')

const viteConfig = read('../../../vite.config.ts')

/** Regexes from `workbox.navigateFallbackDenylist`, as the service worker evaluates them. */
const denylist = (): RegExp[] => {
  const block = /navigateFallbackDenylist:\s*\[([\s\S]*?)\n\s*\],/.exec(viteConfig)?.[1]
  expect(block, 'navigateFallbackDenylist not found in vite.config.ts').toBeTruthy()
  const out: RegExp[] = []
  for (const line of block!.split('\n')) {
    const code = line.replace(/\/\/.*$/, '').trim()
    const m = /^\/(.+)\/([a-z]*),?$/.exec(code)
    if (m) out.push(new RegExp(m[1], m[2]))
  }
  return out
}

/** Dev-server proxy keys: every prefix the UI sends to the API. */
const proxyPrefixes = (): string[] => {
  const block = /proxy:\s*\{([\s\S]*?)\n\s*\},/.exec(viteConfig)?.[1] ?? ''
  return [...block.matchAll(/'(\/[^']+)':/g)].map(m => m[1])
}

/** Top-level paths in `${torrserverHost}/…` of hosts.ts. */
const hostPrefixes = (): string[] => {
  const src = read('../api/hosts.ts')
  return [...new Set([...src.matchAll(/\$\{torrserverHost\}(\/[a-z]+)/g)].map(m => m[1]))]
}

/**
 * Top-level paths registered in server/web/api/route.go, server/web/msx/msx.go and
 * server/web/route.go (static pages excluded). A navigation to any of these must reach
 * the server: Workbox serves index.html for a navigation that is not denied.
 */
const SERVER_ROUTES = [
  '/shutdown',
  '/settings',
  '/waf',
  '/mediabase',
  '/torznab/caps',
  '/torrents',
  '/torrent/upload',
  '/cache',
  '/stream/file.mkv',
  '/play/abc/1',
  '/subs/abc/1',
  '/viewed',
  '/playlistall/all.m3u',
  '/playlist',
  '/playlist/name.m3u',
  '/download/10',
  '/search/query',
  '/storage/settings',
  '/tmdb/settings',
  '/gst/settings',
  '/ffp/status',
  '/ffp/abc/1',
  '/runtime/status',
  '/ssl/status',
  '/echo',
  '/stat',
  '/status',
  '/magnets',
  '/swagger/index.html',
  '/debug/pprof/heap',
  '/msx/start.json',
  '/files',
]

describe('service worker navigateFallbackDenylist', () => {
  const list = denylist()
  const denied = (path: string) => list.some(re => re.test(path))

  it('parses the config', () => {
    expect(list.length).toBeGreaterThan(10)
    expect(denied('/')).toBe(false)
    expect(denied('/index.html')).toBe(false)
  })

  it.each(proxyPrefixes())('denies the dev-proxy prefix %s', prefix => {
    expect(denied(prefix)).toBe(true)
    expect(denied(`${prefix}/x`)).toBe(true)
  })

  it.each(hostPrefixes())('denies the hosts.ts prefix %s', prefix => {
    expect(denied(prefix)).toBe(true)
  })

  it.each(hostPrefixes())('proxies the hosts.ts prefix %s in the dev server', prefix => {
    expect(proxyPrefixes()).toContain(prefix)
  })

  it.each(SERVER_ROUTES)('denies the server route %s', path => {
    expect(denied(path)).toBe(true)
  })
})
