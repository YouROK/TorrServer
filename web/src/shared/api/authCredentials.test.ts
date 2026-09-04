import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import axios from 'axios'

import {
  applyAuthToAxios,
  clearCredentials,
  encodeBasicAuthorization,
  getAuthorizationHeader,
  setCredentials,
  withAuthMediaUrl,
} from './authCredentials'

const memory = new Map<string, string>()

describe('authCredentials', () => {
  beforeEach(() => {
    memory.clear()
    vi.stubGlobal('sessionStorage', {
      getItem: (key: string) => memory.get(key) ?? null,
      setItem: (key: string, value: string) => {
        memory.set(key, value)
      },
      removeItem: (key: string) => {
        memory.delete(key)
      },
      clear: () => memory.clear(),
    })
    vi.stubGlobal('window', {
      location: { origin: 'http://127.0.0.1:8090', protocol: 'http:', pathname: '/' },
      assign: vi.fn(),
    })
  })

  afterEach(() => {
    clearCredentials()
    vi.unstubAllGlobals()
  })

  it('encodes Basic Authorization', () => {
    expect(encodeBasicAuthorization('alice', 's3cret')).toBe(`Basic ${btoa('alice:s3cret')}`)
  })

  it('stores credentials and applies axios default', () => {
    setCredentials('bob', 'hunter2')
    expect(getAuthorizationHeader()).toBe(encodeBasicAuthorization('bob', 'hunter2'))
    expect(axios.defaults.headers.common.Authorization).toBe(encodeBasicAuthorization('bob', 'hunter2'))
    expect(axios.defaults.headers.common['X-Requested-With']).toBe('XMLHttpRequest')
  })

  it('injects userinfo into same-origin media URLs', () => {
    setCredentials('u', 'p')
    const url = withAuthMediaUrl('http://127.0.0.1:8090/stream?link=abc&index=1&play')
    expect(url).toBe('http://u:p@127.0.0.1:8090/stream?link=abc&index=1&play')
  })

  it('applyAuthToAxios sets X-Requested-With without credentials', () => {
    applyAuthToAxios()
    expect(axios.defaults.headers.common.Authorization).toBeUndefined()
    expect(axios.defaults.headers.common['X-Requested-With']).toBe('XMLHttpRequest')
  })
})
