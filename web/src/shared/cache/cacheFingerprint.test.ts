import { describe, expect, it } from 'vitest'
import type { CachePiece } from 'shared/api/types'

import { cheapPiecesFingerprint, cheapReadersFingerprint } from './cacheFingerprint'

describe('cache fingerprints', () => {
  it('mixes 10000 pieces and changes when one size changes', () => {
    const pieces: Record<string, CachePiece> = {}
    for (let i = 0; i < 10_000; i++) {
      pieces[String(i)] = { Size: i % 16, Length: 16, Completed: i % 5 === 0, Priority: i % 6 }
    }
    const before = cheapPiecesFingerprint(pieces)
    pieces['9999'] = { Size: 99, Length: 16, Completed: true, Priority: 1 }
    expect(cheapPiecesFingerprint(pieces)).not.toBe(before)
    expect(cheapPiecesFingerprint(undefined)).toBe(0)
  })

  it('mixes reader ranges including the idle flag', () => {
    const active = cheapReadersFingerprint([{ Reader: 4, Start: 1, End: 8, Active: true }])
    const idle = cheapReadersFingerprint([{ Reader: 4, Start: 1, End: 8, Active: false }])
    expect(active).not.toBe(idle)
    expect(cheapReadersFingerprint(undefined)).toBe(0)
  })
})
