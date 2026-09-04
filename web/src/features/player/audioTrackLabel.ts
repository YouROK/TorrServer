/** GStreamer probe result track — field casing varies by backend build, so lookups are case-insensitive. */
export type ProbeTrack = Record<string, unknown>

export function probeField(track: ProbeTrack, key: string): unknown {
  const lowerKey = key.toLowerCase()
  const match = Object.keys(track).find(candidate => candidate.toLowerCase() === lowerKey)
  return match ? track[match] : undefined
}

export function extractAudioTracks(
  probe: { Tracks?: ProbeTrack[]; tracks?: ProbeTrack[] } | null | undefined,
): ProbeTrack[] {
  return (probe?.Tracks || probe?.tracks || []).filter(
    track => String(probeField(track, 'Type') || '').toLowerCase() === 'audio',
  )
}

export interface AudioTrackDisplay {
  /** Primary line — usually the Title tag (e.g. "DUB | HDRezka Studio"). */
  title: string
  /** Secondary line — LANG · CODEC · N ch · rate. */
  meta: string
}

function formatSampleRate(raw: unknown): string | null {
  const rate = Number(raw)
  if (!Number.isFinite(rate) || rate <= 0) return null
  if (rate >= 1000) return `${Math.round(rate / 1000)} kHz`
  return `${rate} Hz`
}

function formatChannels(raw: unknown): string | null {
  const channels = Number(raw)
  if (!Number.isFinite(channels) || channels <= 0) return null
  return `${channels} ch`
}

const CODEC_ALIASES: Record<string, string> = {
  'audio/x-ac3': 'AC3',
  'audio/ac3': 'AC3',
  ac3: 'AC3',
  'audio/x-eac3': 'E-AC3',
  'audio/eac3': 'E-AC3',
  eac3: 'E-AC3',
  'audio/mpeg': 'AAC',
  'audio/mp4a-latm': 'AAC',
  'audio/x-aac': 'AAC',
  aac: 'AAC',
  'audio/x-opus': 'Opus',
  opus: 'Opus',
  'audio/x-vorbis': 'Vorbis',
  vorbis: 'Vorbis',
  'audio/x-flac': 'FLAC',
  flac: 'FLAC',
  'audio/x-raw': 'PCM',
  'audio/x-wav': 'WAV',
  'audio/x-dts': 'DTS',
  dts: 'DTS',
  'audio/x-true-hd': 'TrueHD',
  truehd: 'TrueHD',
}

/** Strip Gst caps params and map to a short human codec name. Never returns raw `framed=…` strings. */
export function humanizeCodec(raw: string): string {
  const head = raw.split(',')[0]?.trim() || ''
  if (!head) return ''
  const lower = head.toLowerCase()
  if (CODEC_ALIASES[lower]) return CODEC_ALIASES[lower]
  // audio/x-foo → FOO; audio/foo → FOO
  const mimeMatch = lower.match(/^audio\/(?:x-)?(.+)$/)
  if (mimeMatch) {
    const name = mimeMatch[1].replace(/[^a-z0-9+.-]/gi, '')
    if (!name) return ''
    return name.toUpperCase()
  }
  // Already short (AC3, E-AC3) — keep as-is if no commas/params
  if (!raw.includes('=') && !raw.includes('(')) return head
  return head.replace(/^audio\/(?:x-)?/i, '').toUpperCase()
}

/** Two-line label for the GST audio-track picker (title + technical meta). */
export function formatAudioTrackDisplay(track: ProbeTrack, index: number): AudioTrackDisplay {
  const langRaw = String(probeField(track, 'Language') || probeField(track, 'Lang') || '').trim()
  const lang = langRaw ? langRaw.toUpperCase() : ''
  const codec = humanizeCodec(String(probeField(track, 'Codec') || probeField(track, 'CapsName') || '').trim())
  const trackTitle = String(probeField(track, 'Title') || '').trim()
  const channels = formatChannels(probeField(track, 'Channels'))
  const sampleRate = formatSampleRate(probeField(track, 'Rate'))

  const title = trackTitle || lang || codec || `Audio ${index + 1}`
  const metaParts = [title === lang ? '' : lang, codec, channels, sampleRate].filter(Boolean)
  const meta = metaParts.join(' · ')

  return { title, meta }
}
