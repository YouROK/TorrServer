import { describe, expect, it } from 'vitest'
import { IN_DB, WORKING } from 'shared/torrent/states'

import { heroCacheFill } from './heroCacheFill'

const MB = 1024 * 1024

describe('heroCacheFill', () => {
  it('uses the details poll when /cache is not polling, ignoring a stale snapshot', () => {
    const stale = { Filled: 5 * MB, Capacity: 128 * MB }
    expect(heroCacheFill(stale, { stat: WORKING, preloaded_bytes: 9 * MB }, false)).toEqual({
      filled: 9 * MB,
      capacity: 128 * MB,
    })
  })

  it('reads an omitted preloaded_bytes on an active torrent as an empty cache', () => {
    expect(heroCacheFill({}, { stat: WORKING }, false, 64 * MB)).toEqual({ filled: 0, capacity: 64 * MB })
  })

  it('has nothing to show for a torrent with no live cache', () => {
    expect(heroCacheFill({}, { stat: IN_DB }, false, 64 * MB).filled).toBeUndefined()
  })

  it('prefers the live /cache snapshot while it polls', () => {
    const live = { Filled: 7 * MB, Capacity: 128 * MB }
    expect(heroCacheFill(live, { stat: WORKING, preloaded_bytes: 6 * MB }, true, 64 * MB)).toEqual({
      filled: 7 * MB,
      capacity: 128 * MB,
    })
  })
})
