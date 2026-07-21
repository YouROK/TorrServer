import { toApiTorrsLink } from './torrsLink'

export interface LibraryImportItem {
  link: string
  title?: string
  category?: string
  poster?: string
  /** Lowercase infohash when known — used to skip torrents already in the library. */
  hashHint?: string
}

interface ExportJsonRow {
  hash?: string
  title?: string
  name?: string
  category?: string
  poster?: string
  torrs_hash?: string
}

const BTIH_RE = /(?:urn:)?btih:([a-fA-F0-9]{40}|[a-zA-Z2-7]{32})/i
const MAGNET_IN_LINE_RE = /magnet:\?[^\s"'<>]+/i
const TORRS_IN_LINE_RE = /(?:web\+)?torrs:\/\/[^\s"'<>]+/i
const HASH_ONLY_RE = /^\b[0-9a-f]{32}\b$|^\b[0-9a-f]{40}\b$|^\b[0-9a-f]{64}\b$/i
const MAGNET_RE = /^magnet:\?xt=urn:[a-z0-9].*/i
const HTTP_RE = /^(http(s?)):\/\/.*/i
const TORRS_RE = /^(?:web\+)?torrs:\/\/.*/i
const TORRENT_FILE_RE = /^.*\.(torrent)$/i

/** Same acceptance rules as Add dialog sources — kept local to avoid pulling browser hosts into unit tests. */
function isTorrentSource(source: string): boolean {
  return (
    HASH_ONLY_RE.test(source) ||
    MAGNET_RE.test(source) ||
    TORRENT_FILE_RE.test(source) ||
    HTTP_RE.test(source) ||
    TORRS_RE.test(source)
  )
}

function normalizeTorrsLink(value: string): string {
  return toApiTorrsLink(value)
}

function hashFromLink(link: string): string | undefined {
  const match = link.match(BTIH_RE)
  if (match?.[1]) return match[1].toLowerCase()
  if (HASH_ONLY_RE.test(link.trim())) return link.trim().toLowerCase()
  return undefined
}

function magnetForHash(hash: string, title?: string): string {
  const dn = encodeURIComponent(title || hash)
  return `magnet:?xt=urn:btih:${hash}&dn=${dn}`
}

function itemFromJsonRow(row: ExportJsonRow): LibraryImportItem | null {
  const title = (row.title || row.name || '').trim() || undefined
  const category = row.category?.trim() || undefined
  const poster = row.poster?.trim() || undefined

  if (row.torrs_hash?.trim()) {
    const link = normalizeTorrsLink(row.torrs_hash)
    return { link, title, category, poster, hashHint: row.hash?.trim().toLowerCase() || hashFromLink(link) }
  }

  const hash = row.hash?.trim()
  if (hash && HASH_ONLY_RE.test(hash)) {
    return {
      link: magnetForHash(hash, title),
      title,
      category,
      poster,
      hashHint: hash.toLowerCase(),
    }
  }

  return null
}

function itemFromLine(line: string): LibraryImportItem | null {
  const trimmed = line.trim()
  if (!trimmed || trimmed.startsWith('#')) return null

  const magnet = trimmed.match(MAGNET_IN_LINE_RE)?.[0]
  if (magnet && isTorrentSource(magnet)) {
    return { link: magnet, hashHint: hashFromLink(magnet) }
  }

  const torrs = trimmed.match(TORRS_IN_LINE_RE)?.[0]
  if (torrs && isTorrentSource(torrs)) {
    return { link: normalizeTorrsLink(torrs), hashHint: hashFromLink(torrs) }
  }

  if (isTorrentSource(trimmed)) {
    const link = TORRS_RE.test(trimmed) ? normalizeTorrsLink(trimmed) : trimmed
    return { link, hashHint: hashFromLink(trimmed) }
  }

  return null
}

function dedupeItems(items: LibraryImportItem[]): LibraryImportItem[] {
  const seen = new Set<string>()
  const out: LibraryImportItem[] = []
  for (const item of items) {
    const key = (item.hashHint || item.link).toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(item)
  }
  return out
}

/**
 * Parse clipboard / file text produced by Export library (JSON array, magnets,
 * torrs:// / web+torrs:// lines) or a loose mix of the same.
 */
export function parseLibraryImportText(raw: string): LibraryImportItem[] {
  const text = raw.replace(/^\uFEFF/, '').trim()
  if (!text) return []

  if (text.startsWith('[') || text.startsWith('{')) {
    try {
      const parsed: unknown = JSON.parse(text)
      const rows = Array.isArray(parsed) ? parsed : [parsed]
      const items = rows
        .filter((row): row is ExportJsonRow => row != null && typeof row === 'object')
        .map(row => itemFromJsonRow(row as ExportJsonRow))
        .filter((item): item is LibraryImportItem => item != null)
      if (items.length) return dedupeItems(items)
    } catch {
      // Fall through to line parsing — user may have pasted non-JSON that starts with `[`.
    }
  }

  const fromLines = text
    .split(/\r?\n/)
    .map(itemFromLine)
    .filter((item): item is LibraryImportItem => item != null)

  return dedupeItems(fromLines)
}
