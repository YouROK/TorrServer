import axios from 'axios'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('axios', () => ({
  default: { get: vi.fn() },
}))

describe('torrent helpers', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: '127.0.0.1', port: '8090' },
    })
    vi.mocked(axios.get).mockReset()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('accepts hash, magnet, torrent, http, and torrs sources', async () => {
    const { checkTorrentSource } = await import('./torrentHelpers')
    const ok = [
      '0123456789abcdef0123456789abcdef01234567',
      '0123456789abcdef0123456789abcdef',
      '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
      'magnet:?xt=urn:btih:ab',
      'movie.torrent',
      'https://example.test/a.torrent',
      'torrs://ab',
      'web+torrs://ab',
    ]
    for (const source of ok) expect(checkTorrentSource(source)).toBe(true)
    expect(checkTorrentSource('not a torrent')).toBe(false)
    expect(checkTorrentSource('abcd')).toBe(false)
  })

  it('trims release noise before a poster search', async () => {
    const { shortenTitleForPosterSearch } = await import('./torrentHelpers')
    expect(shortenTitleForPosterSearch('')).toBe('')
    expect(shortenTitleForPosterSearch('The Movie [1080p] (2020)')).toBe('The Movie')
    expect(shortenTitleForPosterSearch('Фильм / extra title words here')).toBe('Фильм')
  })

  it('searches TMDB at /3/search/multi when the server base has no version', async () => {
    vi.mocked(axios.get).mockImplementation((url: string) => {
      if (String(url).includes('/tmdb/settings')) {
        return Promise.resolve({
          data: {
            APIKey: 'k',
            APIURL: 'https://api.themoviedb.org',
            ImageURL: 'https://image.tmdb.org',
            ImageURLRu: 'https://imagetmdb.com',
          },
        })
      }
      return Promise.resolve({ data: { results: [{ poster_path: '/p.jpg' }, {}] } })
    })
    const { getMoviePosters } = await import('./torrentHelpers')
    await expect(getMoviePosters('Фильм', 'ru')).resolves.toEqual(['https://imagetmdb.com/t/p/w300/p.jpg'])
    expect(axios.get).toHaveBeenCalledWith(
      'https://api.themoviedb.org/3/search/multi',
      expect.objectContaining({
        params: expect.objectContaining({ api_key: 'k', query: 'Фильм', language: 'ru' }),
      }),
    )
  })
})
