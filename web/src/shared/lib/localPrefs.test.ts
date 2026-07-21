import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { readLocalBool, readLocalJson, writeLocalJson } from './localPrefs'

describe('localPrefs', () => {
  const store = new Map<string, string>()

  beforeEach(() => {
    store.clear()
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
      setItem: (key: string, value: string) => {
        store.set(key, value)
      },
      removeItem: (key: string) => {
        store.delete(key)
      },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('readLocalJson returns fallback for missing / corrupt values', () => {
    expect(readLocalJson('missing', { a: 1 })).toEqual({ a: 1 })
    store.set('bad', '{not-json')
    expect(readLocalJson('bad', [1, 2])).toEqual([1, 2])
  })

  it('readLocalJson parses valid JSON', () => {
    store.set('ok', JSON.stringify({ theme: 'dark' }))
    expect(readLocalJson('ok', { theme: 'light' })).toEqual({ theme: 'dark' })
  })

  it('readLocalBool handles bool / string / number forms', () => {
    store.set('t', 'true')
    expect(readLocalBool('t')).toBe(true)
    store.set('f', JSON.stringify(false))
    expect(readLocalBool('f', true)).toBe(false)
    store.set('one', '1')
    expect(readLocalBool('one')).toBe(true)
  })

  it('writeLocalJson stores stringifyable values', () => {
    writeLocalJson('prefs', { x: 2 })
    expect(store.get('prefs')).toBe('{"x":2}')
  })
})
