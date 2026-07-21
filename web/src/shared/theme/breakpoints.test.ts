import { describe, expect, it } from 'vitest'

import { MEDIA_SHORT_VIEWPORT, MEDIA_TABLET_LANDSCAPE, queryMax } from './breakpoints'

describe('breakpoints', () => {
  it('exports dialog max-width query at 960px', () => {
    expect(queryMax('dialog')).toBe('(max-width: 960px)')
  })

  it('exports tablet landscape media for iPad-style fullscreen sheets', () => {
    expect(MEDIA_TABLET_LANDSCAPE).toContain('orientation: landscape')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('max-width: 1366px')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('max-height: 1024px')
    expect(MEDIA_TABLET_LANDSCAPE).toContain('min-width: 701px')
  })

  it('exports short viewport media', () => {
    expect(MEDIA_SHORT_VIEWPORT).toBe('(max-height: 500px)')
  })
})
