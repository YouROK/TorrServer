import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

describe('settingsEvents', () => {
  beforeEach(() => {
    vi.resetModules()
    const listeners = new Map<string, Set<EventListener>>()
    vi.stubGlobal('window', {
      addEventListener: (type: string, listener: EventListener) => {
        if (!listeners.has(type)) listeners.set(type, new Set())
        listeners.get(type)!.add(listener)
      },
      removeEventListener: (type: string, listener: EventListener) => {
        listeners.get(type)?.delete(listener)
      },
      dispatchEvent: (event: Event) => {
        listeners.get(event.type)?.forEach(listener => listener(event))
        return true
      },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('dispatches SETTINGS_CHANGED_EVENT on window', async () => {
    const { SETTINGS_CHANGED_EVENT, notifySettingsChanged } = await import('./settingsEvents')
    const handler = vi.fn()
    window.addEventListener(SETTINGS_CHANGED_EVENT, handler)
    notifySettingsChanged()
    expect(handler).toHaveBeenCalledTimes(1)
    window.removeEventListener(SETTINGS_CHANGED_EVENT, handler)
  })
})
