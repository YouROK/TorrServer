import { gstSettingsHost } from 'shared/api/hosts'
import { authFetch } from 'shared/api/authCredentials'

export interface GstSettingsResponse {
  built_in?: boolean
  config?: Record<string, unknown>
  defaults?: Record<string, unknown>
  [key: string]: unknown
}

/** Raw fetch (not axios) — GST may be absent; callers handle `built_in: false`. */
export const getGstSettings = async (signal?: AbortSignal): Promise<GstSettingsResponse> => {
  const response = await authFetch(gstSettingsHost(), { signal })
  if (!response.ok) throw new Error('Failed to load GStreamer settings')
  return (await response.json()) as GstSettingsResponse
}

const GST_INT_KEYS = [
  'MaxTasks',
  'InactiveMinutes',
  'AACBitrateKbps',
  'AACChannels',
  'AACSamplerate',
  'SegmentSeconds',
  'SegmentDiff',
  'VideoBitrate',
] as const

/** Go unmarshals these into `int`. A float fails the whole config; NaN becomes JSON null. */
const sanitizeGstConfig = (config: Record<string, unknown>): Record<string, unknown> => {
  const next = { ...config }
  for (const key of GST_INT_KEYS) {
    if (!(key in next)) continue
    const n = Number(next[key])
    next[key] = Number.isFinite(n) ? Math.round(n) : 0
  }
  if ('GSTVersion' in next) {
    const version = Number(next.GSTVersion)
    if (!Number.isFinite(version)) delete next.GSTVersion
  }
  return next
}

export const setGstSettings = async (config: Record<string, unknown>): Promise<void> => {
  const response = await authFetch(gstSettingsHost(), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action: 'set', config: sanitizeGstConfig(config) }),
  })
  if (!response.ok) throw new Error('Failed to save GStreamer settings')
}

/** Reset GST pipeline config to platform defaults on the server (`action: def`). */
export const resetGstSettings = async (): Promise<GstSettingsResponse> => {
  const response = await authFetch(gstSettingsHost(), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action: 'def' }),
  })
  if (!response.ok) throw new Error('Failed to reset GStreamer settings')
  // Server may return empty body — reload so UI gets defaults + current config.
  return getGstSettings()
}
