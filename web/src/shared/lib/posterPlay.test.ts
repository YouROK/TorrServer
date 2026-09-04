import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'

import { buildExternalPlayerHref } from './externalPlayers'

describe('buildExternalPlayerHref', () => {
  const stream = 'https://ts.example/stream/a.mkv?link=h&index=1&play'

  it('builds Infuse / SenPlayer / VLC / IINA deep links', () => {
    const encoded = encodeURIComponent(stream)
    expect(buildExternalPlayerHref('infuse', stream)).toBe(`infuse://x-callback-url/play?url=${encoded}`)
    expect(buildExternalPlayerHref('senPlayer', stream)).toBe(`senplayer://x-callback-url/play?url=${encoded}`)
    expect(buildExternalPlayerHref('vlc', stream)).toBe(`vlc://${stream}`)
    expect(buildExternalPlayerHref('iina', stream)).toBe(`iina://weblink?url=${encoded}`)
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
