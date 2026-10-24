import { QueryClient } from '@tanstack/react-query'
import axios from 'axios'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { IN_DB } from 'shared/torrent/states'

vi.mock('axios', () => ({
  default: {
    post: vi.fn(),
    isAxiosError: (err: unknown) => Boolean((err as { isAxiosError?: boolean })?.isAxiosError),
    isCancel: (err: unknown) => Boolean((err as { isCancel?: boolean })?.isCancel),
  },
}))

describe('torrent API', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: '127.0.0.1', port: '8090' },
    })
    vi.mocked(axios.post).mockReset()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('posts list, get, add, drop, rem, and wipe', async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: { hash: 'ab' } })
    const api = await import('./torrents')

    await api.getTorrents()
    await api.getTorrent('Ab')
    await api.addTorrent({ link: 'magnet:?xt=urn:btih:ab', title: '', poster: '', timestamp: 0 })
    await api.dropTorrent('ab')
    await api.removeTorrent('ab')
    await api.wipeTorrents()

    const bodies = vi.mocked(axios.post).mock.calls.map(call => call[1])
    expect(bodies).toEqual([
      { action: 'list' },
      { action: 'get', hash: 'Ab' },
      {
        action: 'add',
        link: 'magnet:?xt=urn:btih:ab',
        title: undefined,
        category: undefined,
        poster: '',
        save_to_db: true,
        timestamp: undefined,
      },
      { action: 'drop', hash: 'ab' },
      { action: 'rem', hash: 'ab' },
      { action: 'wipe' },
    ])
    expect(String(vi.mocked(axios.post).mock.calls[0][0])).toMatch(/\/torrents$/)
  })

  it('forwards the abort signal and does not wrap a cancelled list', async () => {
    const signal = AbortSignal.abort()
    const cancel = { isCancel: true, message: 'canceled' }
    vi.mocked(axios.post).mockRejectedValue(cancel)
    const { getTorrents } = await import('./torrents')

    await expect(getTorrents(signal)).rejects.toBe(cancel)
    expect(axios.post).toHaveBeenCalledWith(expect.stringMatching(/\/torrents$/), { action: 'list' }, { signal })
  })

  it('maps 401, 403, 500, malformed JSON, and timeouts onto Error.status', async () => {
    const { getTorrents, dropTorrent } = await import('./torrents')
    const cases: [unknown, string, number | undefined][] = [
      [
        { isAxiosError: true, message: 'no', response: { status: 401, data: { error: 'Unauthorized' } } },
        'Unauthorized',
        401,
      ],
      [
        { isAxiosError: true, message: 'no', response: { status: 403, data: { error: 'Read-only mode' } } },
        'Read-only mode',
        403,
      ],
      [
        { isAxiosError: true, message: 'Request failed with status code 500', response: { status: 500, data: {} } },
        'Request failed with status code 500',
        500,
      ],
      [{ isAxiosError: true, message: 'Unexpected end of JSON input' }, 'Unexpected end of JSON input', undefined],
      [
        { isAxiosError: true, message: 'timeout of 10000ms exceeded', code: 'ECONNABORTED' },
        'timeout of 10000ms exceeded',
        undefined,
      ],
      [new Error('network down'), 'network down', undefined],
    ]
    for (const [err, message, status] of cases) {
      vi.mocked(axios.post).mockRejectedValueOnce(err)
      const thrown = await getTorrents().then(
        () => {
          throw new Error('expected rejection')
        },
        (error: Error & { response?: { status?: number } }) => error,
      )
      expect(thrown.message).toBe(message)
      expect(thrown.response?.status).toBe(status)
    }
    vi.mocked(axios.post).mockRejectedValueOnce({
      isAxiosError: true,
      message: 'no',
      response: { status: 403, data: { error: 'Read-only mode' } },
    })
    await expect(dropTorrent('ab')).rejects.toMatchObject({ message: 'Read-only mode', response: { status: 403 } })
  })

  it('does not let an aborted list replace a newer cached row', async () => {
    const { getTorrents, TORRENTS_QUERY_KEY, upsertTorrentsInList, markTorrentsDroppedInList } =
      await import('./torrents')
    const qc = new QueryClient()
    upsertTorrentsInList(qc, { hash: 'Ab', title: 'fresh', stat: 3, download_speed: 9 })
    const signal = AbortSignal.abort()
    vi.mocked(axios.post).mockRejectedValue({ isCancel: true, message: 'canceled' })
    await expect(getTorrents(signal)).rejects.toMatchObject({ isCancel: true })
    expect(qc.getQueryData(TORRENTS_QUERY_KEY)).toEqual([{ hash: 'Ab', title: 'fresh', stat: 3, download_speed: 9 }])

    markTorrentsDroppedInList(qc, 'AB')
    const dropped = qc.getQueryData<[{ stat: number; download_speed: number }]>(TORRENTS_QUERY_KEY)?.[0]
    expect(dropped?.stat).toBe(IN_DB)
    expect(dropped?.download_speed).toBe(0)
  })
})
