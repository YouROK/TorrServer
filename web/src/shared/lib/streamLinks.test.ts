import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'

describe('streamLinks', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubEnv('VITE_SERVER_HOST', '')
    vi.stubGlobal('window', {
      location: { protocol: 'https:', hostname: 'ts.example', port: '8443' },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.unstubAllEnvs()
  })

  it('percent-encodes RFC 3986 extras in the path and keeps play valueless', async () => {
    const { streamPlayUrl, torrentPlaylistUrl } = await import('./streamLinks')
    expect(streamPlayUrl('ab&cd', { id: 1, path: "Season 1/a'b!(*) #?.mkv" })).toBe(
      'https://ts.example:8443/stream/a%27b%21%28%2A%29%20%23%3F.mkv?link=ab%26cd&index=1&play',
    )
    expect(torrentPlaylistUrl('deadbeef', "Show's [2160p]")).toBe(
      'https://ts.example:8443/stream/Show%27s%20%5B2160p%5D.m3u?link=deadbeef&m3u&fn=file.m3u',
    )
  })

  it('builds Infuse, iOS VLC, desktop VLC, IINA, and Android intent links', async () => {
    const { buildAndroidIntentHref, buildExternalPlayerHref } = await import('./streamLinks')
    const stream = 'http://user:p%40ss@192.168.1.5:8090/stream/a%27b.mkv?link=ab%26cd&index=1&play'
    const encoded = encodeURIComponent(stream)

    expect(buildExternalPlayerHref('infuse', stream)).toBe(`infuse://x-callback-url/play?url=${encoded}`)
    expect(buildExternalPlayerHref('iina', stream)).toBe(`iina://weblink?url=${encoded}`)
    expect(buildExternalPlayerHref('vlc', stream)).toBe(`vlc://${stream}`)
    expect(buildExternalPlayerHref('vlc', stream, { ios: true })).toBe(
      `vlc-x-callback://x-callback-url/stream?url=${encoded}`,
    )
    expect(buildExternalPlayerHref('vlc', stream, { android: true })).toBe(buildAndroidIntentHref(stream))
    expect(buildAndroidIntentHref(stream)).toBe(
      'intent://user:p%40ss@192.168.1.5:8090/stream/a%27b.mkv?link=ab%26cd&index=1&play#Intent;scheme=http;type=video/*;end',
    )
  })
})
