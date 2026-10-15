import { afterEach, describe, expect, it, vi } from 'vitest'

import { bindPlaylistButton, bindPlaylistClick, playlistFilename, savePlaylistInApp } from './openBrowser'

describe('playlist in-app save', () => {
  it('only saves inside an iOS home-screen app', () => {
    expect(savePlaylistInApp(true, true)).toBe(true)
    expect(savePlaylistInApp(true, false)).toBe(false)
    expect(savePlaylistInApp(false, true)).toBe(false)
  })

  it('keeps an m3u name from the path', () => {
    expect(playlistFilename('/stream/My%20Show.m3u')).toBe('My Show.m3u')
    expect(playlistFilename('/playlistall/all.m3u')).toBe('all.m3u')
    expect(playlistFilename('/stream/')).toBe('playlist.m3u')
  })

  it('cancels the click in an iOS home-screen app so the page stays put', () => {
    vi.stubGlobal('navigator', { userAgent: 'iPhone', platform: 'iPhone', maxTouchPoints: 1 })
    vi.stubGlobal('window', {
      matchMedia: () => ({ matches: true }),
      navigator: { standalone: true },
      location: { origin: 'http://127.0.0.1:8090' },
    })
    const preventDefault = vi.fn()
    bindPlaylistClick('http://127.0.0.1:8090/stream/Show.m3u', () => {})({ preventDefault })
    expect(preventDefault).toHaveBeenCalledOnce()
  })

  it('opens an external player instead of the m3u', () => {
    const location = { href: '' }
    vi.stubGlobal('window', { location })
    const preventDefault = vi.fn()
    bindPlaylistButton(
      'vlc',
      'http://127.0.0.1:8090/stream/Show.m3u',
      'vlc://http://127.0.0.1:8090/stream/Show.m3u',
      () => {},
    )({
      preventDefault,
    })
    expect(preventDefault).toHaveBeenCalledOnce()
    expect(location.href).toBe('vlc://http://127.0.0.1:8090/stream/Show.m3u')
  })

  it('copies the playlist url when the action is copyLink', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('navigator', { clipboard: { writeText } })
    const preventDefault = vi.fn()
    const onResult = vi.fn()
    bindPlaylistButton('copyLink', 'http://127.0.0.1:8090/stream/Show.m3u', null, onResult)({ preventDefault })
    expect(preventDefault).toHaveBeenCalledOnce()
    await vi.waitFor(() => expect(onResult).toHaveBeenCalledWith('copied'))
    expect(writeText).toHaveBeenCalledWith('http://127.0.0.1:8090/stream/Show.m3u')
  })
})

afterEach(() => {
  vi.unstubAllGlobals()
})
