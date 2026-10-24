import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import axios from 'axios'

vi.mock('axios', () => ({
  default: { post: vi.fn() },
}))

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

  it('maps empty, NaN, and negative inputs onto wire types Go can unmarshal', async () => {
    const { sanitizeBTSetsForWire } = await import('./settings')
    const out = sanitizeBTSetsForWire({
      CacheSize: '' as unknown as number,
      ReaderReadAHead: -3,
      PreloadCache: -1,
      ConnectionsLimit: Number.NaN,
      DownloadRateLimit: 'nope' as unknown as number,
      UploadRateLimit: -5,
      SslPort: '' as unknown as number,
      UseDisk: false,
      TorrentsSavePath: '',
      EnableDLNA: true,
      FriendlyName: 'Box',
      DefaultTrackers: 'udp://x',
      StoreViewedInJson: true,
      StoreSettingsInJson: false,
    })
    // Number('') is 0. The server rewrites CacheSize 0 to 64MiB; the client must not send a float.
    expect(out.CacheSize).toBe(0)
    expect(out.ReaderReadAHead).toBe(5)
    expect(out.PreloadCache).toBe(0)
    expect(out.ConnectionsLimit).toBe(25)
    expect(out.DownloadRateLimit).toBe(0)
    expect(out.UploadRateLimit).toBe(-5)
    expect(out.SslPort).toBe(0)
    expect(out.UseDisk).toBe(false)
    expect(out.EnableDLNA).toBe(true)
    expect(out.FriendlyName).toBe('Box')
    expect(out.DefaultTrackers).toBe('udp://x')
    expect(out.StoreViewedInJson).toBe(true)
    expect(out.StoreSettingsInJson).toBe(false)
    for (const key of [
      'CacheSize',
      'ReaderReadAHead',
      'PreloadCache',
      'ConnectionsLimit',
      'DownloadRateLimit',
      'UploadRateLimit',
      'SslPort',
    ]) {
      expect(typeof out[key]).toBe('number')
      expect(Number.isInteger(out[key])).toBe(true)
    }
  })

  it('posts sanitized settings and leaves a 403 read-only error untouched', async () => {
    const { getSettings, setSettings } = await import('./settings')
    const signal = new AbortController().signal
    vi.mocked(axios.post).mockResolvedValueOnce({ data: { CacheSize: 64 } })
    await expect(getSettings(signal)).resolves.toEqual({ CacheSize: 64 })
    expect(axios.post).toHaveBeenCalledWith(expect.stringMatching(/\/settings$/), { action: 'get' }, { signal })

    vi.mocked(axios.post).mockResolvedValueOnce({ data: null })
    await setSettings({ CacheSize: 10.2, ReaderReadAHead: 95.4, StoreViewedInJson: true, EnableDLNA: false })
    const body = vi.mocked(axios.post).mock.calls.at(-1)?.[1] as {
      action: string
      sets: { CacheSize: number; ReaderReadAHead: number; StoreViewedInJson: boolean; EnableDLNA: boolean }
    }
    expect(body.action).toBe('set')
    expect(body.sets.CacheSize).toBe(10)
    expect(body.sets.ReaderReadAHead).toBe(95)
    expect(body.sets.StoreViewedInJson).toBe(true)
    expect(body.sets.EnableDLNA).toBe(false)

    vi.mocked(axios.post).mockRejectedValueOnce({ response: { status: 403, data: { error: 'Read-only mode' } } })
    await expect(setSettings({ CacheSize: 1 })).rejects.toMatchObject({ response: { status: 403 } })
  })
})
