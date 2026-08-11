import i18n from 'shared/i18n'
import type { CachePiece, TorrentCache, TorrentStat } from 'shared/api/types'
import { forEachPiece, isReaderActive } from 'shared/cache/buildCacheMap'

/** Human file/cache size using binary 1024 steps and localized unit labels. */
export function humanizeSize(size?: number | null): string {
  if (size == null || Number.isNaN(size) || size < 0) return '—'
  if (size === 0) return `0 ${i18n.t('B')}`
  const i = Math.floor(Math.log(size) / Math.log(1024))
  return `${Number((size / Math.pow(1024, i)).toFixed(2))} ${
    [i18n.t('B'), i18n.t('KB'), i18n.t('MB'), i18n.t('GB'), i18n.t('TB')][i]
  }`
}

/**
 * Torrent transfer rate for the UI.
 * Server reports **bytes/s**; we multiply by 8 and format with SI 1000 (bps/kbps/Mbps).
 */
export function humanizeSpeed(speed?: number | null): string {
  if (speed == null || Number.isNaN(speed) || speed < 0) return `0 ${i18n.t('bps')}`
  if (speed === 0) return `0 ${i18n.t('bps')}`
  const i = Math.floor(Math.log(speed * 8) / Math.log(1000))
  return `${Number(((speed * 8) / Math.pow(1000, i)).toFixed(0))} ${
    [i18n.t('bps'), i18n.t('kbps'), i18n.t('Mbps'), i18n.t('Gbps'), i18n.t('Tbps')][i]
  }`
}

/** `active/total · seeders` peer summary, or null when torrent is missing. */
export function getPeerString(torrent?: TorrentStat | null): string | null {
  if (!torrent) return null
  const active = torrent.active_peers
  const total = torrent.total_peers
  if (active == null) return '—'
  const seeders = torrent.connected_seeders ?? 0
  return `${active}/${total ?? 0} · ${seeders}`
}

/**
 * Cache fill label — same shape as Details «Кеш» widget.
 * Shows raw `filled` (may exceed capacity until cleanPieces). Percent is capped
 * at 100 so a brief overfill never reads as "103% capacity".
 */
export function formatCacheFilledLabel(
  filled?: number | null,
  capacity?: number | null,
  opts?: { percent?: 'whenOver' | 'always' },
): string | null {
  if (filled == null || capacity == null || capacity <= 0 || filled < 0) return null
  const pct = Math.min(100, Math.round((filled / capacity) * 100))
  const over = filled > capacity
  const sizes = `${humanizeSize(filled)} / ${humanizeSize(capacity)}`
  const mode = opts?.percent ?? 'whenOver'
  if (mode === 'always' || over) return `${sizes} · ${pct}%`
  return sizes
}

/** Preload target: CacheSize × PreloadCache% — matches server `Preload()` in apihelper.go. */
export function resolveBufferTargetBytes(capacity?: number | null, preloadCachePercent?: number | null): number | null {
  if (capacity == null || capacity <= 0) return null
  const pct = preloadCachePercent == null || Number.isNaN(preloadCachePercent) ? 50 : preloadCachePercent
  const preloadSize = (capacity / 100) * Math.max(0, pct)
  if (preloadSize <= 0) return null
  // Never exceed Cache Size — same clamp as the server.
  return Math.min(preloadSize, capacity)
}

/**
 * Preload progress: live Filled toward the Preload target (not Cache Size capacity).
 * Label order is `Filled / preloadTarget`. Percent is capped at 100.
 * Use only for idle/preload — while streaming prefer {@link formatBufferAheadLabel}.
 */
export function formatBufferFilledLabel(
  filled?: number | null,
  bufferTarget?: number | null,
  opts?: { percent?: 'whenOver' | 'always' },
): string | null {
  if (filled == null || bufferTarget == null || bufferTarget <= 0 || filled < 0) return null
  const rawPct = Math.round((filled / bufferTarget) * 100)
  const pct = Math.min(100, rawPct)
  const over = filled > bufferTarget
  const sizes = `${humanizeSize(filled)} / ${humanizeSize(bufferTarget)}`
  const mode = opts?.percent ?? 'whenOver'
  if (mode === 'always' || over) return `${sizes} · ${pct}%`
  return sizes
}

/**
 * Playable contiguous bytes ahead of the playhead — never `ahead / target`, which
 * reads as if the preload target were a capacity that overflowed.
 * Pass `0` for the empty-ahead label (`BufferAheadEmpty`).
 */
export function formatBufferAheadLabel(aheadBytes?: number | null): string | null {
  if (aheadBytes == null || aheadBytes < 0 || Number.isNaN(aheadBytes)) return null
  if (aheadBytes === 0) return i18n.t('BufferAheadEmpty')
  return i18n.t('BufferAheadLabel', { size: humanizeSize(aheadBytes) })
}

export function bufferFillPercent(filled?: number | null, bufferTarget?: number | null): number {
  if (filled == null || bufferTarget == null || bufferTarget <= 0 || filled < 0) return 0
  return Math.min(100, Math.max(0, (filled / bufferTarget) * 100))
}

/**
 * Contiguous ready bytes from the active playhead forward — how much can be
 * played before the stream stalls. Total `Filled` pins at 100% during playback
 * and stops being informative; this does not.
 * Returns null when no active reader exists (preload or idle) so callers can
 * fall back to `Filled`.
 */
export function bufferAheadBytes(cache?: TorrentCache | null): number | null {
  const piecesCount = cache?.PiecesCount ?? 0
  if (!cache || piecesCount <= 0) return null

  let head = -1
  for (const reader of cache.Readers ?? []) {
    if (!isReaderActive(reader)) continue
    const piece = reader.Reader
    if (piece != null && piece >= 0 && piece < piecesCount && piece > head) head = piece
  }
  if (head < 0) return null

  const pieceById = new Map<number, CachePiece>()
  forEachPiece(cache.Pieces, (id, piece) => {
    if (id >= head) pieceById.set(id, piece)
  })

  const defaultLength = cache.PiecesLength || 0
  let total = 0
  for (let id = head; id < piecesCount; id++) {
    const piece = pieceById.get(id)
    if (!piece) break
    const length = piece.Length || defaultLength || 0
    if (length <= 0) break
    const size = Math.min(piece.Size || 0, length)
    // Count the partial head/hole piece, then stop — bytes past a hole are not playable.
    total += size
    if (size < length) break
  }
  return total
}

/**
 * Strip unbalanced trailing bracket groups and trailing dots/spaces from titles
 * (common in release-group naming) so display names stay readable.
 */
export const removeRedundantCharacters = (string: string): string => {
  let newString = string
  const brackets: Array<[string, string]> = [
    ['(', ')'],
    ['[', ']'],
    ['{', '}'],
  ]

  brackets.forEach(el => {
    const leftBracketRegex = new RegExp(`\\${el[0]}`, 'g')
    const leftBracketAmount = [...newString.matchAll(leftBracketRegex)].length
    const rightBracketRegex = new RegExp(`\\${el[1]}`, 'g')
    const rightBracketAmount = [...newString.matchAll(rightBracketRegex)].length

    if (leftBracketAmount !== rightBracketAmount) {
      const removeRegex = new RegExp(`(\\${el[0]})(?!.*\\1).*`, 'g')
      newString = newString.replace(removeRegex, '')
    }
  })

  const hasThreeDotsAtTheEnd = !!newString.match(/\.{3}$/g)
  const trimmedString = newString.replace(/[\\.| ]+$/g, '').trim()

  return hasThreeDotsAtTheEnd ? `${trimmedString}..` : trimmedString
}

/** Binary (1024) size with fixed English unit labels — prefer {@link humanizeSize} in UI. */
export function formatSizeToClassicUnits(bytes?: number | null): string {
  if (!bytes || bytes === 0) return '0 B'
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const value = bytes / Math.pow(1024, i)
  return `${value.toFixed(i === 0 ? 0 : 2)} ${sizes[i]}`
}

/**
 * Parse Torznab/search size strings into bytes.
 * `KiB`/`MiB`/`…iB` (or `CiB`) use base 1024; plain `KB`/`MB` use decimal 1000.
 */
export function parseSizeToBytes(sizeStr?: string | null): number {
  if (!sizeStr || typeof sizeStr !== 'string') return 0

  if (sizeStr.trim().match(/^\d+\s*B$/i)) {
    const digits = sizeStr.trim().match(/^\d+/)
    return digits ? parseInt(digits[0], 10) : 0
  }

  const match = sizeStr.trim().match(/^([\d.]+)\s*([KMGT]?)(i?B|CiB)$/i)
  if (!match) return 0

  const value = parseFloat(match[1])
  const unit = match[2].toUpperCase()
  const suffix = match[3].toUpperCase()

  if (Number.isNaN(value)) return 0

  const isBinary = suffix.includes('I') || suffix.includes('C')
  const base = isBinary ? 1024 : 1000
  const multipliers: Record<string, number> = { '': 1, K: 1, M: 2, G: 3, T: 4 }
  const multiplier = multipliers[unit] || 1

  return Math.round(value * Math.pow(base, multiplier))
}
