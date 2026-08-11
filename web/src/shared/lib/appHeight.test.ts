import { describe, expect, it } from 'vitest'

import { preferVisibleViewportHeight } from './appHeight'

describe('preferVisibleViewportHeight', () => {
  it('prefers visualViewport when present (Safari tab visible band)', () => {
    // Large clientHeight (≈100vh) must not win over a shorter visual viewport.
    expect(preferVisibleViewportHeight(700, 844, 700)).toBe(700)
    expect(preferVisibleViewportHeight(650, 900, 650)).toBe(650)
  })

  it('falls back to innerHeight when visual is missing', () => {
    expect(preferVisibleViewportHeight(720, 900, 0)).toBe(720)
  })

  it('falls back to clientHeight only when both visual and inner are missing', () => {
    expect(preferVisibleViewportHeight(0, 800, 0)).toBe(800)
  })
})
