import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const getTorrent = vi.fn()

vi.mock('shared/api/torrents', () => ({
  getTorrent: (...args: unknown[]) => getTorrent(...args),
}))

import { waitForPlayableFiles } from './waitForPlayableFiles'

describe('waitForPlayableFiles', () => {
  beforeEach(() => {
    getTorrent.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns live file_stats when already present', async () => {
    getTorrent.mockResolvedValue({
      file_stats: [{ id: 1, path: 'movie.mkv', length: 10 }],
    })

    const result = await waitForPlayableFiles('abc', { timeoutMs: 1000, intervalMs: 50 })
    expect(result.playable).toHaveLength(1)
    expect(result.playable[0].path).toBe('movie.mkv')
    expect(getTorrent).toHaveBeenCalledTimes(1)
  })

  it('uses metadataData immediately when live stats are empty', async () => {
    const data = JSON.stringify({
      TorrServer: { Files: [{ id: 2, path: 'film.mp4', length: 20 }] },
    })
    getTorrent.mockResolvedValue({ file_stats: [], data: '' })

    const result = await waitForPlayableFiles('abc', { metadataData: data, timeoutMs: 1000 })
    expect(result.playable).toHaveLength(1)
    expect(result.playable[0].id).toBe(2)
  })

  it('polls until file_stats appear', async () => {
    vi.useFakeTimers()
    getTorrent
      .mockResolvedValueOnce({ file_stats: [] })
      .mockResolvedValueOnce({ file_stats: [] })
      .mockResolvedValueOnce({ file_stats: [{ id: 3, path: 'ep.mkv', length: 1 }] })

    const promise = waitForPlayableFiles('abc', { timeoutMs: 5000, intervalMs: 100 })
    await vi.advanceTimersByTimeAsync(250)
    const result = await promise
    expect(result.playable).toHaveLength(1)
    expect(getTorrent.mock.calls.length).toBeGreaterThanOrEqual(3)
  })

  it('returns empty after timeout', async () => {
    vi.useFakeTimers()
    getTorrent.mockResolvedValue({ file_stats: [] })

    const promise = waitForPlayableFiles('abc', { timeoutMs: 300, intervalMs: 100 })
    await vi.advanceTimersByTimeAsync(400)
    const result = await promise
    expect(result.playable).toEqual([])
  })
})
