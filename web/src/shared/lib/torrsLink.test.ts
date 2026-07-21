import { describe, expect, it } from 'vitest'

import { isTorrsLink, toApiTorrsLink, toShareTorrsLink } from './torrsLink'

describe('torrsLink', () => {
  it('detects torrs and web+torrs schemes', () => {
    expect(isTorrsLink('torrs://abc')).toBe(true)
    expect(isTorrsLink('web+torrs://abc')).toBe(true)
    expect(isTorrsLink('magnet:?xt=urn:btih:aaa')).toBe(false)
  })

  it('normalizes to API torrs://', () => {
    expect(toApiTorrsLink('web+torrs://packed')).toBe('torrs://packed')
    expect(toApiTorrsLink('torrs://packed')).toBe('torrs://packed')
    expect(toApiTorrsLink('packed')).toBe('torrs://packed')
    expect(toApiTorrsLink('/packed')).toBe('torrs://packed')
  })

  it('builds share links as web+torrs://', () => {
    expect(toShareTorrsLink('torrs://packed')).toBe('web+torrs://packed')
    expect(toShareTorrsLink('web+torrs://packed')).toBe('web+torrs://packed')
    expect(toShareTorrsLink('packed')).toBe('web+torrs://packed')
  })
})
