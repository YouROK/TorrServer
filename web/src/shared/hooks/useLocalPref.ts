import { useCallback, useEffect, useState } from 'react'

import { readLocalBool, readLocalJson, writeLocalJson } from 'shared/lib/localPrefs'

const LOCAL_PREF_EVENT = 'torrserver:local-pref'

/** React-friendly localStorage bool that updates across the app when prefs change. */
export function useLocalBoolPref(key: string, fallback = false): [boolean, (next: boolean) => void] {
  const [value, setValue] = useState(() => readLocalBool(key, fallback))

  useEffect(() => {
    const sync = () => setValue(readLocalBool(key, fallback))
    const onStorage = (e: StorageEvent) => {
      if (e.key === key) sync()
    }
    const onCustom = (e: Event) => {
      const detail = (e as CustomEvent<{ key?: string }>).detail
      if (!detail?.key || detail.key === key) sync()
    }
    window.addEventListener('storage', onStorage)
    window.addEventListener(LOCAL_PREF_EVENT, onCustom as EventListener)
    return () => {
      window.removeEventListener('storage', onStorage)
      window.removeEventListener(LOCAL_PREF_EVENT, onCustom as EventListener)
    }
  }, [key, fallback])

  const set = useCallback(
    (next: boolean) => {
      writeLocalJson(key, next)
      setValue(next)
      window.dispatchEvent(new CustomEvent(LOCAL_PREF_EVENT, { detail: { key } }))
    },
    [key],
  )

  return [value, set]
}

/**
 * Typed JSON localStorage preference with the same cross-tab / in-app sync as
 * {@link useLocalBoolPref} (`storage` + `torrserver:local-pref` custom event).
 */
export function useLocalJsonPref<T>(key: string, fallback: T): [T, (next: T) => void] {
  const [value, setValue] = useState(() => readLocalJson(key, fallback))

  useEffect(() => {
    const sync = () => setValue(readLocalJson(key, fallback))
    const onStorage = (e: StorageEvent) => {
      if (e.key === key) sync()
    }
    const onCustom = (e: Event) => {
      const detail = (e as CustomEvent<{ key?: string }>).detail
      if (!detail?.key || detail.key === key) sync()
    }
    window.addEventListener('storage', onStorage)
    window.addEventListener(LOCAL_PREF_EVENT, onCustom as EventListener)
    return () => {
      window.removeEventListener('storage', onStorage)
      window.removeEventListener(LOCAL_PREF_EVENT, onCustom as EventListener)
    }
  }, [key, fallback])

  const set = useCallback(
    (next: T) => {
      writeLocalJson(key, next)
      setValue(next)
      window.dispatchEvent(new CustomEvent(LOCAL_PREF_EVENT, { detail: { key } }))
    },
    [key],
  )

  return [value, set]
}

export { LOCAL_PREF_EVENT }
