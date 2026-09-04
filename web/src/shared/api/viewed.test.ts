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
})
