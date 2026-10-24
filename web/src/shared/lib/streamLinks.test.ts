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

  it('encodes spaces, Cyrillic, brackets, quotes, and query characters in the path', async () => {
    const { encodePathSegment, streamPlayUrl, streamPreloadUrl, torrentPlaylistUrl } = await import('./streamLinks')
    const names: [string, string][] = [
      ['a b.mkv', 'a%20b.mkv'],
      ['фильм.mkv', '%D1%84%D0%B8%D0%BB%D1%8C%D0%BC.mkv'],
      ['a[1080p].mkv', 'a%5B1080p%5D.mkv'],
      ['a(2020).mkv', 'a%282020%29.mkv'],
      ["a'b.mkv", 'a%27b.mkv'],
      ['a#b.mkv', 'a%23b.mkv'],
      ['a?b.mkv', 'a%3Fb.mkv'],
      ['a&amp.mkv', 'a%26amp.mkv'],
      ['a%b.mkv', 'a%25b.mkv'],
      ['a!(*).mkv', 'a%21%28%2A%29.mkv'],
    ]
    for (const [name, encoded] of names) {
      expect(encodePathSegment(name)).toBe(encoded)
      const url = new URL(streamPlayUrl('hash&id#1', { id: 7, path: `Season 1/${name}` }))
      expect(url.pathname).toBe(`/stream/${encoded}`)
      expect(url.search).toBe('?link=hash%26id%231&index=7&play')
      expect(decodeURIComponent(url.pathname.slice('/stream/'.length))).toBe(name)
    }
    expect(streamPreloadUrl('dead beef', 3)).toBe('https://ts.example:8443/stream?link=dead+beef&index=3&preload')
    expect(torrentPlaylistUrl('ab', 'Show [2160p]')).toBe(
      'https://ts.example:8443/stream/Show%20%5B2160p%5D.m3u?link=ab&m3u&fn=file.m3u',
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
    expect(buildExternalPlayerHref('senPlayer', stream)).toBe(`senplayer://x-callback-url/play?url=${encoded}`)
    expect(buildAndroidIntentHref(stream.replace('http:', 'https:'))).toContain('#Intent;scheme=https;')
  })

  it('falls back to one m3u url when an Infuse episode list exceeds the scheme limit', async () => {
    const { INFUSE_PLAYLIST_HREF_LIMIT, buildExternalPlayerHref, buildInfusePlaylistHref } =
      await import('./streamLinks')
    const short = buildInfusePlaylistHref(
      [{ url: 'http://ts/stream/a.mkv?play', filename: 'a.mkv' }],
      'http://ts/show.m3u',
    )
    expect(short.startsWith('infuse://x-callback-url/play?url=')).toBe(true)
    expect(short).toContain(encodeURIComponent('a.mkv'))
    expect(short.length).toBeLessThanOrEqual(INFUSE_PLAYLIST_HREF_LIMIT)

    const episodes = Array.from({ length: 80 }, (_, i) => ({
      url: `http://192.168.1.50:8090/stream/${'episode'.repeat(8)}-${i}.mkv?link=ab&index=${i}&play`,
      filename: `episode-${i}.mkv`,
    }))
    const m3u = 'http://192.168.1.50:8090/stream/Show.m3u?link=ab&m3u'
    expect(buildInfusePlaylistHref(episodes, m3u)).toBe(buildExternalPlayerHref('infuse', m3u))
  })
})
