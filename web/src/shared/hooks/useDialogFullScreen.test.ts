import { describe, expect, it } from 'vitest'

import { resolveDialogFullScreen } from './useDialogFullScreen'

describe('resolveDialogFullScreen', () => {
  it('is on for phones / narrow tablets (≤960)', () => {
    expect(resolveDialogFullScreen({ narrow: true, tabletLandscape: false, force: false })).toBe(true)
  })

  it('is on for tablet landscape without the Appearance pref', () => {
    expect(resolveDialogFullScreen({ narrow: false, tabletLandscape: true, force: false })).toBe(true)
  })

  it('is off on wide monitors when the Appearance pref is off', () => {
    expect(resolveDialogFullScreen({ narrow: false, tabletLandscape: false, force: false })).toBe(false)
  })

  it('forces fullscreen on wide monitors when the Appearance pref is on', () => {
    expect(resolveDialogFullScreen({ narrow: false, tabletLandscape: false, force: true })).toBe(true)
  })

  it('stays on when both auto rules and force apply', () => {
    expect(resolveDialogFullScreen({ narrow: true, tabletLandscape: true, force: true })).toBe(true)
  })
})
