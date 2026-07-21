import type { FfpProbeResult, FfpStream } from 'shared/api/extras'

/** Format ffprobe duration (seconds) as `m:ss` or `h:mm:ss`. */
export function formatFfpDuration(secondsRaw?: string | number | null): string | null {
  if (secondsRaw == null || secondsRaw === '') return null
  const total = Math.round(Number(secondsRaw))
  if (!Number.isFinite(total) || total < 0) return null
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

/** Bitrate string from ffprobe (bits/s) → Mbps/kbps display. */
export function formatFfpBitrate(raw?: string | null): string | null {
  if (!raw) return null
  const bits = Number(raw)
  if (!Number.isFinite(bits) || bits <= 0) return raw
  const mbps = bits / 1_000_000
  if (mbps >= 1) {
    const rounded = mbps >= 10 ? mbps.toFixed(0) : Number(mbps.toFixed(1)).toString()
    return `${rounded} Mbps`
  }
  const kbps = bits / 1000
  return `${kbps.toFixed(0)} kbps`
}

export function formatFfpBytes(raw?: string | null): string | null {
  if (!raw) return null
  const bytes = Number(raw)
  if (!Number.isFinite(bytes) || bytes < 0) return raw
  if (bytes >= 1024 ** 3) return `${(bytes / 1024 ** 3).toFixed(2)} GB`
  if (bytes >= 1024 ** 2) return `${(bytes / 1024 ** 2).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${bytes} B`
}

/**
 * Read a stream tag by exact key or `key-*` prefixed variant (ffprobe language/title tags).
 */
export function streamTag(stream: FfpStream, key: string): string {
  const tags = stream.tags
  if (!tags) return ''
  const direct = tags[key]
  if (typeof direct === 'string' && direct) return direct
  const found = Object.entries(tags).find(([k]) => k === key || k.startsWith(`${key}-`))
  return found && typeof found[1] === 'string' ? found[1] : ''
}

/** Partition ffprobe streams for MediaInfo sections. */
export function groupStreams(data: FfpProbeResult | null) {
  const streams = data?.streams ?? []
  return {
    video: streams.filter(s => s.codec_type === 'video'),
    audio: streams.filter(s => s.codec_type === 'audio'),
    subtitle: streams.filter(s => s.codec_type === 'subtitle'),
    other: streams.filter(s => s.codec_type !== 'video' && s.codec_type !== 'audio' && s.codec_type !== 'subtitle'),
  }
}

/** Convert `r_frame_rate` (`num/den`) to a display FPS string. */
export function fpsFromRate(rate?: string | null): string | null {
  if (!rate || rate === '0/0') return null
  const [a, b] = rate.split('/').map(Number)
  if (!Number.isFinite(a) || !Number.isFinite(b) || b === 0) return rate
  const fps = a / b
  return Number.isInteger(fps) ? `${fps}` : fps.toFixed(2)
}
