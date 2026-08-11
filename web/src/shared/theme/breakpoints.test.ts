import { describe, expect, it } from 'vitest'

import { MEDIA_SHORT_VIEWPORT, MEDIA_TABLET_LANDSCAPE, queryMax, queryMin } from './breakpoints'

describe('breakpoints', () => {
  it('exports dialog max-width query at 960px', () => {
    expect(queryMax('dialog')).toBe('(max-width: 960px)')
  })

  it('exports phone min-width query for 6-chip hero band', () => {
    expect(queryMin('phone')).toBe('(min-width: 420px)')
  })

  it('exports tablet landscape media for iPad-style fullscreen sheets', () => {
    expect(MEDIA_TABLET_LANDSCAPE).toContain('orientation: landscape')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('max-width: 1366px')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('max-height: 1024px')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('min-width: 701px')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('hover: none')
  })

  it('exports short viewport media', () => {
    expect(MEDIA_SHORT_VIEWPORT).toBe('(max-height: 500px)')
  })
})
