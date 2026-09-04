import axios from 'axios'
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { playbackDevicesHost, playbackPlayHost } from 'utils/Hosts'

export const LOCAL_PLAYBACK_DEVICE_ID = 'this-device'
export const PLAYBACK_CLIENT_MODES = {
  CONTROL_ONLY: 'control-only',
  CONTROL_AND_PLAYBACK: 'control-and-playback',
}
export const PLAYBACK_ROUTING_MODES = {
  LOCAL: 'local',
  PRIMARY: 'primary',
  PER_BROWSER: 'per-browser',
}

const MODE_STORAGE_KEY = 'playbackClientMode'
const TARGET_STORAGE_KEY = 'playbackTargetId'

const PlaybackContext = createContext(null)

const defaultConfig = {
  enabled: false,
  routing_mode: PLAYBACK_ROUTING_MODES.LOCAL,
  primary_device_id: '',
  devices: [],
}

const storedMode = () => {
  const value = localStorage.getItem(MODE_STORAGE_KEY)
  return Object.values(PLAYBACK_CLIENT_MODES).includes(value) ? value : PLAYBACK_CLIENT_MODES.CONTROL_AND_PLAYBACK
}

const normalizeConfig = data => {
  if (Array.isArray(data)) return { ...defaultConfig, devices: data }
  return {
    ...defaultConfig,
    ...(data || {}),
    devices: Array.isArray(data?.devices) ? data.devices : [],
  }
}

export const PlaybackProvider = ({ children }) => {
  const { t } = useTranslation()
  const [config, setConfig] = useState(defaultConfig)
  const [isLoadingDevices, setIsLoadingDevices] = useState(true)
  const [mode, setModeState] = useState(storedMode)
  const [targetId, setTargetIdState] = useState(localStorage.getItem(TARGET_STORAGE_KEY) || '')

  const refreshPlaybackConfig = useCallback(async () => {
    setIsLoadingDevices(true)
    try {
      const { data } = await axios.get(playbackDevicesHost())
      setConfig(normalizeConfig(data))
    } finally {
      setIsLoadingDevices(false)
    }
  }, [])

  useEffect(() => {
    refreshPlaybackConfig().catch(() => setConfig(defaultConfig))
  }, [refreshPlaybackConfig])

  const localTarget = useMemo(
    () => ({ id: LOCAL_PLAYBACK_DEVICE_ID, name: t('Playback.ThisDevice'), isLocal: true }),
    [t],
  )
  const remoteTargets = useMemo(() => config.devices.map(device => ({ ...device, isLocal: false })), [config.devices])

  const targets = useMemo(() => {
    if (!config.enabled || config.routing_mode === PLAYBACK_ROUTING_MODES.LOCAL) return [localTarget]
    if (config.routing_mode === PLAYBACK_ROUTING_MODES.PRIMARY) {
      const primary = remoteTargets.find(device => device.id === config.primary_device_id)
      return primary ? [primary] : []
    }
    if (mode === PLAYBACK_CLIENT_MODES.CONTROL_ONLY) return remoteTargets
    return [localTarget, ...remoteTargets]
  }, [config.enabled, config.primary_device_id, config.routing_mode, localTarget, mode, remoteTargets])

  useEffect(() => {
    if (config.routing_mode !== PLAYBACK_ROUTING_MODES.PER_BROWSER || !config.enabled) return
    if (targets.some(target => target.id === targetId)) return
    const fallback = targets[0]?.id || ''
    setTargetIdState(fallback)
    if (fallback) localStorage.setItem(TARGET_STORAGE_KEY, fallback)
    else localStorage.removeItem(TARGET_STORAGE_KEY)
  }, [config.enabled, config.routing_mode, targetId, targets])

  const setMode = useCallback(value => {
    if (!Object.values(PLAYBACK_CLIENT_MODES).includes(value)) return
    setModeState(value)
    localStorage.setItem(MODE_STORAGE_KEY, value)
  }, [])

  const setTargetId = useCallback(value => {
    setTargetIdState(value)
    if (value) localStorage.setItem(TARGET_STORAGE_KEY, value)
    else localStorage.removeItem(TARGET_STORAGE_KEY)
  }, [])

  const selectedTarget = useMemo(() => {
    if (!config.enabled || config.routing_mode === PLAYBACK_ROUTING_MODES.LOCAL) return localTarget
    if (config.routing_mode === PLAYBACK_ROUTING_MODES.PRIMARY) return targets[0] || null
    return targets.find(target => target.id === targetId) || null
  }, [config.enabled, config.routing_mode, localTarget, targetId, targets])

  const launchVlc = useCallback(
    async ({ path, hash, index, streamUrl }) => {
      if (!selectedTarget) throw new Error('playback target is not selected')

      if (selectedTarget.isLocal) {
        const absoluteStreamUrl = new URL(streamUrl, window.location.href)
        window.location.href = `vlc://${absoluteStreamUrl}`
        return
      }

      await axios.post(playbackPlayHost(), {
        device_id: selectedTarget.id,
        path,
        hash,
        index,
      })
    },
    [selectedTarget],
  )

  const value = useMemo(
    () => ({
      config,
      devices: config.devices,
      isLoadingDevices,
      launchVlc,
      mode,
      refreshPlaybackConfig,
      remotePlaybackEnabled: config.enabled,
      routingMode: config.routing_mode,
      selectedTarget,
      setMode,
      setTargetId,
      showTargetSelector: config.enabled && config.routing_mode === PLAYBACK_ROUTING_MODES.PER_BROWSER,
      targetId,
      targets,
    }),
    [
      config,
      isLoadingDevices,
      launchVlc,
      mode,
      refreshPlaybackConfig,
      selectedTarget,
      setMode,
      setTargetId,
      targetId,
      targets,
    ],
  )

  return <PlaybackContext.Provider value={value}>{children}</PlaybackContext.Provider>
}

export const usePlayback = () => {
  const context = useContext(PlaybackContext)
  if (!context) throw new Error('usePlayback must be used inside PlaybackProvider')
  return context
}
