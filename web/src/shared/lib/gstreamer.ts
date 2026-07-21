import { useQuery } from '@tanstack/react-query'

import type { GStreamerRuntime } from 'shared/api/types'
import { authFetch } from 'shared/api/authCredentials'
import { getTorrServerHost, gstSettingsHost } from 'shared/api/hosts'

export const GST_RUNTIME_QUERY_KEY = 'gstreamer-runtime-settings'

const unavailableRuntime: GStreamerRuntime = { built_in: false }

const loadGStreamerRuntime = async (): Promise<GStreamerRuntime> => {
  const response = await authFetch(gstSettingsHost())
  if (!response.ok) return unavailableRuntime
  return response.json()
}

/**
 * Cached GST capability/config. Soft-fails to `{ built_in: false }` when the
 * endpoint is missing or errors — non-GST builds must still boot the UI.
 */
export const useGStreamerRuntime = (): GStreamerRuntime => {
  const { data } = useQuery<GStreamerRuntime, Error>({
    queryKey: [GST_RUNTIME_QUERY_KEY],
    queryFn: loadGStreamerRuntime,
    staleTime: 60 * 1000,
    gcTime: 5 * 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  })

  return data || unavailableRuntime
}

const fileExtension = (path: string): string => {
  const fileName = path.split('?')[0]
  const dot = fileName.lastIndexOf('.')
  return dot === -1 ? '' : fileName.slice(dot + 1).toLowerCase()
}

/**
 * Whether this path should open via the GStreamer HLS path instead of progressive stream.
 * MKV/WebM always when GST is built in; AVI only when `TranscodeAVI` is enabled.
 * Broader allowlist still applies for Play button visibility — see `isFilePlayable`.
 */
export const shouldUseGStreamerPlayer = (path: string, runtime?: GStreamerRuntime | null): boolean => {
  if (!runtime?.built_in) return false

  switch (fileExtension(path)) {
    case 'mkv':
    case 'mk3d':
    case 'webm':
      return true
    case 'avi':
      return Boolean(runtime.config?.TranscodeAVI)
    default:
      return false
  }
}

export const gstreamerMasterUrl = (hash: string, fileID: string | number, audio = 0): string =>
  `${getTorrServerHost()}/gst/${encodeURIComponent(hash)}/master.m3u8?index=${encodeURIComponent(
    String(fileID),
  )}&audio=${encodeURIComponent(String(audio))}`

export const gstreamerProbeUrl = (hash: string, fileID: string | number): string =>
  `${getTorrServerHost()}/gst/${encodeURIComponent(hash)}/probe?index=${encodeURIComponent(String(fileID))}`

export const gstreamerHeartbeatUrl = (hash: string): string =>
  `${getTorrServerHost()}/gst/${encodeURIComponent(hash)}/heartbeat`
