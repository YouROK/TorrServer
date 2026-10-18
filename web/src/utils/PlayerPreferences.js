import { useEffect, useState } from 'react'

const PLAYER_SETTINGS_EVENT = 'torrserver-player-settings-changed'

const readBooleanPreference = (key, defaultValue = false) => {
  const value = localStorage.getItem(key)
  return value === null ? defaultValue : JSON.parse(value)
}

export const notifyPlayerSettingsChanged = () => {
  window.dispatchEvent(new Event(PLAYER_SETTINGS_EVENT))
}

export const useBooleanPlayerPreference = (key, defaultValue = false) => {
  const [value, setValue] = useState(() => readBooleanPreference(key, defaultValue))

  useEffect(() => {
    const update = event => {
      if (event.type === 'storage' && event.key && event.key !== key) return
      setValue(readBooleanPreference(key, defaultValue))
    }

    window.addEventListener(PLAYER_SETTINGS_EVENT, update)
    window.addEventListener('storage', update)
    return () => {
      window.removeEventListener(PLAYER_SETTINGS_EVENT, update)
      window.removeEventListener('storage', update)
    }
  }, [defaultValue, key])

  return value
}
