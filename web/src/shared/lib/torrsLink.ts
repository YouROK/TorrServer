/**
 * Shared `torrs://` / `web+torrs://` helpers.
 * Clipboard / PWA protocol_handlers use `web+torrs://`; the TorrServer API expects `torrs://`.
 */

const WEB_TORRS_RE = /^web\+torrs:\/\//i
const TORRS_RE = /^torrs:\/\//i

/** True for `torrs://…` or `web+torrs://…`. */
export function isTorrsLink(value: string): boolean {
  const trimmed = value.trim()
  return WEB_TORRS_RE.test(trimmed) || TORRS_RE.test(trimmed)
}

/**
 * Normalize a packed token / torrs / web+torrs payload into `torrs://…` for addTorrent.
 * Leaves non-torrs strings unchanged (caller should only pass torrs-like values when unsure).
 */
export function toApiTorrsLink(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return trimmed
  if (WEB_TORRS_RE.test(trimmed)) return trimmed.replace(/^web\+/i, '')
  if (TORRS_RE.test(trimmed)) return trimmed
  return `torrs://${trimmed.replace(/^\/+/, '')}`
}

/** Clipboard / export form that Chromium PWA protocol_handlers will open. */
export function toShareTorrsLink(raw: string): string {
  const api = toApiTorrsLink(raw)
  if (!api) return api
  return api.replace(/^torrs:\/\//i, 'web+torrs://')
}
