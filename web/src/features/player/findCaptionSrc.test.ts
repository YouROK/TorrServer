import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('shared/i18n', () => ({
  default: { t: (key: string) => key },
}))

import type { TorrentFileStat } from 'shared/api/types'

describe('findCaptionSrc', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: 'localhost', port: '8090' },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('returns a stream URL for a matching sidecar subtitle', async () => {
    const { findCaptionSrc } = await import('./usePlayLauncher')
    const video = { id: 0, path: 'Season 1/episode.mkv', length: 100 }
    const files: TorrentFileStat[] = [video, { id: 1, path: 'Season 1/episode.en.srt', length: 1 }]

    expect(findCaptionSrc(video, files, 'deadbeef')).toBe(
      'http://localhost:8090/stream/episode.en.srt?link=deadbeef&index=1&play',
    )
  })

  it('returns empty string when no caption file matches', async () => {
    const { findCaptionSrc } = await import('./usePlayLauncher')
    const video = { id: 0, path: 'movie.mkv', length: 100 }
    const files: TorrentFileStat[] = [video, { id: 1, path: 'readme.txt', length: 1 }]

    expect(findCaptionSrc(video, files, 'hash')).toBe('')
  })

  it('accepts Path/Id field casing on torrent files', async () => {
    const { findCaptionSrc } = await import('./usePlayLauncher')
    const video = { Id: 2, Path: 'film.mp4', Length: 100 }
    const files: TorrentFileStat[] = [video, { Id: 3, Path: 'film.vtt', Length: 1 }]

    expect(findCaptionSrc({ id: 2, path: 'film.mp4', length: 100 }, files, 'abc')).toContain('film.vtt')
  })
})
