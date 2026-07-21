import { toApiTorrsLink } from './torrsLink'

/**
 * Normalize a `?torrs=` protocol-handler payload into a full `torrs://…` link.
 * Chromium may pass the full URL, a bare token/hash, or a `web+torrs://` form.
 */
export function normalizeTorrsLaunchParam(raw: string): string {
  return toApiTorrsLink(raw)
}

/** Pull the first `torrs://…` / `web+torrs://…` token out of shared text / a free-form URL param. */
export function extractTorrsFromText(text: string): string | null {
  const match = text.match(/(?:web\+)?torrs:\/\/[^\s]+/i)
  return match ? toApiTorrsLink(match[0]) : null
}

/**
 * Read launch payload from the current page URL (protocol_handlers / share_target).
 * Clears the query string after a successful read so a refresh does not re-open Add.
 */
export function readLaunchSourceFromUrl(
  search = typeof window !== 'undefined' ? window.location.search : '',
): string | null {
  const params = new URLSearchParams(search)
  const torrs = params.get('torrs')
  if (torrs != null && torrs !== '') {
    return normalizeTorrsLaunchParam(torrs)
  }

  const source = params.get('magnet') || params.get('url') || params.get('text')
  if (!source) return null

  const embeddedTorrs = extractTorrsFromText(source)
  return embeddedTorrs || source
}

/** Whether the current URL carries a launch payload that should be consumed once. */
export function hasLaunchQuery(search = typeof window !== 'undefined' ? window.location.search : ''): boolean {
  const params = new URLSearchParams(search)
  return Boolean(params.get('torrs') || params.get('magnet') || params.get('url') || params.get('text'))
}
