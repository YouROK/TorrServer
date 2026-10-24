import axios from 'axios'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('axios', () => ({
  default: { post: vi.fn() },
}))

describe('getCache', () => {
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

  it('posts action get and passes the abort signal', async () => {
    const signal = new AbortController().signal
    vi.mocked(axios.post).mockResolvedValue({ data: { PiecesCount: 4, Filled: 1 } })
    const { getCache } = await import('./cache')
    await expect(getCache('ab', signal)).resolves.toEqual({ PiecesCount: 4, Filled: 1 })
    expect(axios.post).toHaveBeenCalledWith(
      expect.stringMatching(/\/cache$/),
      { action: 'get', hash: 'ab' },
      { signal },
    )
  })

  it('returns an empty snapshot for a null body and rethrows 500 and cancel', async () => {
    const { getCache } = await import('./cache')
    vi.mocked(axios.post).mockResolvedValueOnce({ data: null })
    await expect(getCache('ab')).resolves.toEqual({})

    vi.mocked(axios.post).mockRejectedValueOnce(Object.assign(new Error('status 500'), { response: { status: 500 } }))
    await expect(getCache('ab')).rejects.toThrow('status 500')

    const signal = AbortSignal.abort()
    vi.mocked(axios.post).mockRejectedValueOnce(Object.assign(new Error('canceled'), { code: 'ERR_CANCELED' }))
    await expect(getCache('ab', signal)).rejects.toThrow('canceled')
  })
})
