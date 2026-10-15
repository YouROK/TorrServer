import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'

import { buildExternalPlayerHref, buildInfusePlaylistHref, playlistActionHref } from './externalPlayers'

describe('buildExternalPlayerHref', () => {
  const stream = 'https://ts.example/stream/a.mkv?link=h&index=1&play'

  it('builds Infuse / SenPlayer / VLC / IINA deep links', () => {
    const encoded = encodeURIComponent(stream)
    expect(buildExternalPlayerHref('infuse', stream)).toBe(`infuse://x-callback-url/play?url=${encoded}`)
    expect(buildExternalPlayerHref('senPlayer', stream)).toBe(`senplayer://x-callback-url/play?url=${encoded}`)
    expect(buildExternalPlayerHref('vlc', stream)).toBe(`vlc://${stream}`)
    expect(playlistActionHref('download', stream)).toBeNull()
    expect(playlistActionHref('copyLink', stream)).toBeNull()
    expect(playlistActionHref('vlc', stream)).toBe(`vlc://${stream}`)
    expect(buildExternalPlayerHref('iina', stream)).toBe(`iina://weblink?url=${encoded}`)
  })

  it('builds an Infuse episode list and sends other players the m3u', () => {
    const m3u = 'https://ts.example/stream/Show.m3u?link=h&m3u'
    const episodes = [
      { url: 'https://ts.example/stream/a.mkv?link=h&index=1&play', filename: 'a.mkv' },
      { url: 'https://ts.example/stream/b.mkv?link=h&index=2&play', filename: 'b.mkv' },
    ]
    const href = buildInfusePlaylistHref(episodes, m3u)
    expect(href).toBe(
      'infuse://x-callback-url/play?' +
        episodes
          .map(episode => `url=${encodeURIComponent(episode.url)}&filename=${encodeURIComponent(episode.filename)}`)
          .join('&'),
    )
    expect(buildExternalPlayerHref('iina', m3u)).toBe(`iina://weblink?url=${encodeURIComponent(m3u)}`)
    expect(buildExternalPlayerHref('senPlayer', m3u)).toBe(
      `senplayer://x-callback-url/play?url=${encodeURIComponent(m3u)}`,
    )
    expect(buildExternalPlayerHref('vlc', m3u)).toBe(`vlc://${m3u}`)

    const long = Array.from({ length: 80 }, (_, index) => ({
      url: `https://ts.example/stream/${'episode'.repeat(8)}-${index}.mkv?link=h&index=${index}&play`,
      filename: `${'episode'.repeat(8)}-${index}.mkv`,
    }))
    expect(buildInfusePlaylistHref(long, m3u)).toBe(buildExternalPlayerHref('infuse', m3u))
  })
})

describe('posterPlay', () => {
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

  it('defaults to copyLink on iOS and builtin elsewhere', async () => {
    const { defaultPosterPlayAction } = await import('./posterPlay')
    expect(defaultPosterPlayAction(true)).toBe('copyLink')
    expect(defaultPosterPlayAction(false)).toBe('builtin')
  })

  it('validates known action strings', async () => {
    const { isPosterPlayAction } = await import('./posterPlay')
    expect(isPosterPlayAction('builtin')).toBe(true)
    expect(isPosterPlayAction('copyLink')).toBe(true)
    expect(isPosterPlayAction('infuse')).toBe(true)
    expect(isPosterPlayAction('nope')).toBe(false)
    expect(isPosterPlayAction(null)).toBe(false)
  })

  it('coerces invalid or disabled external prefs to the platform default', async () => {
    const { coercePosterPlayAction } = await import('./posterPlay')
    const base = {
      isVlcUsed: false,
      isInfuseUsed: false,
      isSenPlayerUsed: false,
      isIinaUsed: false,
      isApple: false,
      isMac: false,
    }
    expect(coercePosterPlayAction('garbage', base, true)).toBe('copyLink')
    expect(coercePosterPlayAction('garbage', base, false)).toBe('builtin')
    expect(coercePosterPlayAction('infuse', { ...base, isApple: true, isInfuseUsed: false }, true)).toBe('copyLink')
    expect(coercePosterPlayAction('infuse', { ...base, isApple: true, isInfuseUsed: true }, true)).toBe('infuse')
    expect(coercePosterPlayAction('vlc', { ...base, isVlcUsed: true }, false)).toBe('vlc')
    expect(coercePosterPlayAction('iina', { ...base, isMac: true, isIinaUsed: false }, false)).toBe('builtin')
  })

  it('keeps builtin and copyLink regardless of player toggles', async () => {
    const { coercePosterPlayAction } = await import('./posterPlay')
    const base = {
      isVlcUsed: false,
      isInfuseUsed: false,
      isSenPlayerUsed: false,
      isIinaUsed: false,
      isApple: false,
      isMac: false,
    }
    expect(coercePosterPlayAction('builtin', base, true)).toBe('builtin')
    expect(coercePosterPlayAction('copyLink', base, false)).toBe('copyLink')
  })

  it('coerces playlist button prefs back to download when the player is off', async () => {
    const { availablePlaylistButtonActions, coercePlaylistButtonAction } = await import('./posterPlay')
    const base = {
      isVlcUsed: false,
      isInfuseUsed: false,
      isSenPlayerUsed: false,
      isIinaUsed: false,
      isApple: false,
      isMac: false,
    }
    expect(coercePlaylistButtonAction('garbage', base)).toBe('download')
    expect(coercePlaylistButtonAction('vlc', base)).toBe('download')
    expect(coercePlaylistButtonAction('copyLink', base)).toBe('copyLink')
    expect(coercePlaylistButtonAction('vlc', { ...base, isVlcUsed: true })).toBe('vlc')
    expect(coercePlaylistButtonAction('infuse', { ...base, isApple: true, isInfuseUsed: false })).toBe('download')
    expect(availablePlaylistButtonActions({ ...base, isApple: true, isInfuseUsed: true })).toEqual([
      'download',
      'copyLink',
      'infuse',
    ])
  })

  it('lists only enabled platform-available externals in settings options', async () => {
    const { availablePosterPlayActions, isExternalPosterPlayEnabled } = await import('./posterPlay')
    const base = {
      isVlcUsed: false,
      isInfuseUsed: false,
      isSenPlayerUsed: false,
      isIinaUsed: false,
      isApple: false,
      isMac: false,
    }
    expect(availablePosterPlayActions(base)).toEqual(['builtin', 'copyLink'])
    expect(
      availablePosterPlayActions({
        ...base,
        isVlcUsed: true,
        isApple: true,
        isInfuseUsed: true,
        isMac: true,
        isIinaUsed: true,
      }),
    ).toEqual(['builtin', 'copyLink', 'vlc', 'infuse', 'iina'])
    expect(isExternalPosterPlayEnabled('senPlayer', { ...base, isApple: true, isSenPlayerUsed: true })).toBe(true)
    expect(isExternalPosterPlayEnabled('senPlayer', { ...base, isApple: false, isSenPlayerUsed: true })).toBe(false)
  })

  it('builds magnet, stream, and playlist URLs', async () => {
    const { magnetFromHash, streamPlayUrl, torrentPlaylistUrl } = await import('./posterPlay')
    expect(magnetFromHash('abc123', 'Show & Title')).toBe('magnet:?xt=urn:btih:abc123&dn=Show%20%26%20Title')
    expect(streamPlayUrl('deadbeef', { id: 2, path: 'folder/video.mkv' })).toBe(
      'https://ts.example:8443/stream/video.mkv?link=deadbeef&index=2&play',
    )
    expect(torrentPlaylistUrl('deadbeef', 'My Show')).toBe(
      'https://ts.example:8443/stream/My%20Show.m3u?link=deadbeef&m3u',
    )
  })
})
