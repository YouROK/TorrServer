import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import axios from 'axios'

vi.mock('axios', () => ({
  default: { post: vi.fn() },
}))

describe('listViewedFiles', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: 'localhost', port: '8090' },
    })
    vi.mocked(axios.post).mockReset()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
  })

  it('returns undefined when the server sends no entries', async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: [] })
    const { listViewedFiles } = await import('./viewed')
    await expect(listViewedFiles('abc123')).resolves.toBeUndefined()
  })

  it('returns sorted file indexes from viewed rows', async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: [{ file_index: 3 }, { file_index: 1 }, { file_index: 2 }],
    })
    const { listViewedFiles } = await import('./viewed')
    await expect(listViewedFiles('abc123')).resolves.toEqual([1, 2, 3])
  })

  it('treats non-array responses as empty', async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: null })
    const { listViewedFiles } = await import('./viewed')
    await expect(listViewedFiles('abc123')).resolves.toBeUndefined()
  })

  it('posts set, rem, and clear with the viewed payload', async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: null })
    const { clearViewedFiles, remViewedFile, setViewedFile } = await import('./viewed')
    await setViewedFile('abc', 2, 15)
    await remViewedFile('abc', 2)
    await clearViewedFiles('abc')
    expect(vi.mocked(axios.post).mock.calls.map(call => call[1])).toEqual([
      { action: 'set', hash: 'abc', file_index: 2, timecode: 15 },
      { action: 'rem', hash: 'abc', file_index: 2 },
      { action: 'rem', hash: 'abc', file_index: -1 },
    ])
  })

  it('rejects when viewed returns 500', async () => {
    vi.mocked(axios.post).mockRejectedValue(Object.assign(new Error('status 500'), { response: { status: 500 } }))
    const { listViewedEntries } = await import('./viewed')
    await expect(listViewedEntries('abc')).rejects.toThrow('status 500')
  })
})
