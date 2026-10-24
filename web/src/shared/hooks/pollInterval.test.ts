import { describe, expect, it } from 'vitest'

import { GETTING_INFO, IN_DB, PRELOAD, WORKING } from 'shared/torrent/states'

import { torrentDetailRefetchIntervalMs, torrentsRefetchIntervalMs } from './pollInterval'

describe('poll intervals', () => {
  it('slows the library poll while the tab is hidden and speeds up metadata resolve', () => {
    expect(torrentsRefetchIntervalMs(true, [{ hash: 'a', stat: GETTING_INFO }])).toBe(10_000)
    expect(torrentsRefetchIntervalMs(false, [{ hash: 'a', stat: GETTING_INFO }])).toBe(500)
    expect(torrentsRefetchIntervalMs(false, [{ hash: 'a', stat: PRELOAD }])).toBe(1000)
    expect(torrentsRefetchIntervalMs(false, [{ hash: 'a', stat: WORKING }])).toBe(1000)
    expect(torrentsRefetchIntervalMs(false, undefined)).toBe(1000)
  })

  it('slows detail polling while hidden and follows GETTING_INFO / PRELOAD / idle', () => {
    expect(torrentDetailRefetchIntervalMs(true, GETTING_INFO)).toBe(5000)
    expect(torrentDetailRefetchIntervalMs(false, GETTING_INFO)).toBe(500)
    expect(torrentDetailRefetchIntervalMs(false, PRELOAD)).toBe(1000)
    expect(torrentDetailRefetchIntervalMs(false, WORKING)).toBe(2000)
    expect(torrentDetailRefetchIntervalMs(false, IN_DB)).toBe(2000)
    expect(torrentDetailRefetchIntervalMs(false, undefined)).toBe(2000)
  })
})
