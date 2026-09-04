import { describe, expect, it } from 'vitest'

import { hasTelegramMiniAppQuery, needsTelegramSdkState } from './telegramWebApp'

describe('telegramWebApp', () => {
  it('detects Mini App query tg=1', () => {
    expect(hasTelegramMiniAppQuery('?tg=1')).toBe(true)
    expect(hasTelegramMiniAppQuery('?foo=1&tg=1')).toBe(true)
    expect(hasTelegramMiniAppQuery('?tg=0')).toBe(false)
    expect(hasTelegramMiniAppQuery('')).toBe(false)
  })

  it('needsTelegramSdkState when query has tg=1', () => {
    expect(needsTelegramSdkState('?tg=1', false)).toBe(true)
  })

  it('needsTelegramSdkState when WebApp already present', () => {
    expect(needsTelegramSdkState('', true)).toBe(true)
  })

  it('does not need SDK for normal browsing', () => {
    expect(needsTelegramSdkState('', false)).toBe(false)
    expect(needsTelegramSdkState('?foo=bar', false)).toBe(false)
  })
})
