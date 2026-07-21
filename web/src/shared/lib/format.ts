import i18n from 'shared/i18n'
import type { TorrentStat } from 'shared/api/types'

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
 * `percent: 'whenOver'` → sizes only until filled exceeds capacity, then `· N%` (can exceed 100).
 * `percent: 'always'` → always append `· N%` (card progress).
 */
export function formatCacheFilledLabel(
  filled?: number | null,
  capacity?: number | null,
  opts?: { percent?: 'whenOver' | 'always' },
): string | null {
  if (filled == null || capacity == null || capacity <= 0 || filled < 0) return null
  const shown = Math.min(filled, capacity)
  const pct = Math.round((filled / capacity) * 100)
  const over = filled > capacity
  const sizes = `${humanizeSize(shown)} / ${humanizeSize(capacity)}`
  const mode = opts?.percent ?? 'whenOver'
  if (mode === 'always' || over) return `${sizes} · ${pct}%`
  return sizes
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
