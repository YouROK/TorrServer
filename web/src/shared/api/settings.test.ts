import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

describe('sanitizeBTSetsForWire', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: '127.0.0.1', port: '8090' },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('rounds floats and replaces NaN so Go can unmarshal ints', async () => {
    const { sanitizeBTSetsForWire } = await import('./settings')
    const out = sanitizeBTSetsForWire({
      CacheSize: 64.2 * 1024 * 1024,
      ReaderReadAHead: 95.6,
      PreloadCache: Number.NaN,
      ConnectionsLimit: Number.NaN,
      RetrackersMode: '2',
      DownloadRateLimit: 1.2,
      PeersListenPort: 200,
    })
    expect(out.CacheSize).toBe(Math.round(64.2 * 1024 * 1024))
    expect(out.ReaderReadAHead).toBe(96)
    expect(out.PreloadCache).toBe(50)
    expect(out.ConnectionsLimit).toBe(25)
    expect(out.RetrackersMode).toBe(2)
    expect(out.DownloadRateLimit).toBe(1)
    expect(out.PeersListenPort).toBe(200)
    expect(sanitizeBTSetsForWire({ ReaderReadAHead: 1, PreloadCache: 140 }).ReaderReadAHead).toBe(5)
    expect(sanitizeBTSetsForWire({ ReaderReadAHead: 1, PreloadCache: 140 }).PreloadCache).toBe(100)
  })
})
