import { describe, expect, it } from 'vitest'

import { beginSearchRequest, isCurrentSearch } from './searchRequest'

describe('searchRequest helpers', () => {
  it('beginSearchRequest aborts previous controller and bumps generation', () => {
    const abortRef = { current: null as AbortController | null }
    const genRef = { current: 0 }

    const first = beginSearchRequest(abortRef, genRef)
    expect(first.gen).toBe(1)
    expect(abortRef.current).toBe(first.ac)
    expect(first.ac.signal.aborted).toBe(false)

    const second = beginSearchRequest(abortRef, genRef)
    expect(first.ac.signal.aborted).toBe(true)
    expect(second.gen).toBe(2)
    expect(abortRef.current).toBe(second.ac)
    expect(isCurrentSearch(first.gen, genRef.current, first.ac.signal)).toBe(false)
    expect(isCurrentSearch(second.gen, genRef.current, second.ac.signal)).toBe(true)
  })

  it('isCurrentSearch rejects aborted or stale generations', () => {
    const ac = new AbortController()
    expect(isCurrentSearch(3, 3, ac.signal)).toBe(true)
    expect(isCurrentSearch(2, 3, ac.signal)).toBe(false)
    ac.abort()
    expect(isCurrentSearch(3, 3, ac.signal)).toBe(false)
  })
})
