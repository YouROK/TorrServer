import { describe, expect, it } from 'vitest'

import {
  extractTorrsFromText,
  hasLaunchQuery,
  normalizeTorrsLaunchParam,
  readLaunchSourceFromUrl,
} from './launchSource'

describe('launchSource', () => {
  it('normalizes torrs protocol-handler payloads', () => {
    expect(normalizeTorrsLaunchParam('torrs://abc')).toBe('torrs://abc')
    expect(normalizeTorrsLaunchParam('web+torrs://abc')).toBe('torrs://abc')
    expect(normalizeTorrsLaunchParam('2c864150aa1f5b60134257bf41eb651b8be1acb7')).toBe(
      'torrs://2c864150aa1f5b60134257bf41eb651b8be1acb7',
    )
    expect(normalizeTorrsLaunchParam('/packedToken')).toBe('torrs://packedToken')
  })

  it('extracts torrs:// from shared text', () => {
    expect(extractTorrsFromText('see torrs://deadbeef and more')).toBe('torrs://deadbeef')
    expect(extractTorrsFromText('see web+torrs://deadbeef and more')).toBe('torrs://deadbeef')
    expect(extractTorrsFromText('magnet:?xt=urn:btih:aaa')).toBeNull()
  })

  it('reads ?torrs= ahead of magnet/url/text', () => {
    expect(readLaunchSourceFromUrl('?torrs=2c864150aa1f5b60134257bf41eb651b8be1acb7&magnet=x')).toBe(
      'torrs://2c864150aa1f5b60134257bf41eb651b8be1acb7',
    )
    expect(readLaunchSourceFromUrl('?magnet=magnet:?xt=urn:btih:aaa')).toBe('magnet:?xt=urn:btih:aaa')
    expect(readLaunchSourceFromUrl('?text=please%20add%20torrs://packedtoken')).toBe('torrs://packedtoken')
    expect(readLaunchSourceFromUrl('?url=https://example.com')).toBe('https://example.com')
    expect(readLaunchSourceFromUrl('')).toBeNull()
  })

  it('detects launch query keys', () => {
    expect(hasLaunchQuery('?torrs=abc')).toBe(true)
    expect(hasLaunchQuery('?magnet=x')).toBe(true)
    expect(hasLaunchQuery('?foo=1')).toBe(false)
  })
})
