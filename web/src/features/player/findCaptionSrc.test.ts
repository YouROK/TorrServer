import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

describe('resolveCaptionSrc', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: 'localhost', port: '8090' },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('returns absolute /subs track URL preferring matching lang', async () => {
    vi.doMock('shared/api/authCredentials', () => ({
      authFetch: vi.fn().mockResolvedValue({
        ok: true,
        json: async () => [
          { index: 2, name: 'ep.en.srt', title: 'en', lang: 'en', format: 'srt', url: '/subs/deadbeef/1?track=2' },
          { index: 3, name: 'ep.ru.ass', title: 'ru', lang: 'ru', format: 'ass', url: '/subs/deadbeef/1?track=3' },
        ],
      }),
    }))
    const { resolveCaptionSrc } = await import('./findCaptionSrc')
    expect(await resolveCaptionSrc('deadbeef', 1, 'ru-RU')).toBe('http://localhost:8090/subs/deadbeef/1?track=3')
  })

  it('returns empty string when API has no tracks', async () => {
    vi.doMock('shared/api/authCredentials', () => ({
      authFetch: vi.fn().mockResolvedValue({ ok: true, json: async () => [] }),
    }))
    const { resolveCaptionSrc } = await import('./findCaptionSrc')
    expect(await resolveCaptionSrc('hash', 0)).toBe('')
  })

  it('returns empty string on fetch failure', async () => {
    vi.doMock('shared/api/authCredentials', () => ({
      authFetch: vi.fn().mockRejectedValue(new Error('network')),
    }))
    const { resolveCaptionSrc } = await import('./findCaptionSrc')
    expect(await resolveCaptionSrc('hash', 0)).toBe('')
  })
})
