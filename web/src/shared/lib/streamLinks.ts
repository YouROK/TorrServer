import { playlistTorrHost, streamHost } from 'shared/api/hosts'

import type { ExternalPosterPlayAction, PlaylistButtonAction } from './posterPlay'

/**
 * RFC 3986 path segment. `encodeURIComponent` leaves `!'()*` raw; those break
 * Infuse, VLC, and IINA when a torrent name contains them.
 */
export const encodePathSegment = (segment: string): string =>
  encodeURIComponent(segment).replace(/[!'()*]/g, ch => `%${ch.charCodeAt(0).toString(16).toUpperCase()}`)

const fileBaseName = (path: string): string => path.split('\\').pop()?.split('/').pop() || path

/** Query string for `/stream`. Flag keys (`play`, `m3u`) stay valueless — Gin `GetQuery` only checks presence. */
export const streamQuery = (hash: string, params: Record<string, string>, flags: readonly string[]): string => {
  const query = new URLSearchParams({ link: hash, ...params }).toString()
  return flags.length ? `${query}&${flags.join('&')}` : query
}

/** Direct play stream URL. Path segment and `link` are encoded (`#`, `?`, `&`, `%`, `'`, `!`, non-ASCII). */
export const streamPlayUrl = (hash: string, file: { id: number; path: string }): string =>
  `${streamHost()}/${encodePathSegment(fileBaseName(file.path))}?${streamQuery(hash, { index: String(file.id) }, ['play'])}`

/** Buffer a file without opening it (`GET /stream?…&preload`). */
export const streamPreloadUrl = (hash: string, fileId: number): string =>
  `${streamHost()}?${streamQuery(hash, { index: String(fileId) }, ['preload'])}`

/**
 * M3U playlist URL for the whole torrent.
 * `fn=file.m3u` matches server `all.m3u` entries — ForkPlayer mishandles a link whose path ends in `.m3u`.
 * Basic userinfo is applied later by `externalLink` / `withAuthMediaUrl`, and only for the page host.
 */
export const torrentPlaylistUrl = (hash: string, displayName: string): string =>
  `${playlistTorrHost()}/${encodePathSegment(displayName)}.m3u?${streamQuery(hash, {}, ['m3u', 'fn=file.m3u'])}`

export interface PlayerLinkPlatform {
  ios?: boolean
  android?: boolean
}

/**
 * Android view intent. `absoluteUrl` keeps userinfo and the encoded path;
 * the scheme moves into the Intent fragment so `#` in the stream URL cannot appear
 * (path encoding already percent-encodes it).
 */
export const buildAndroidIntentHref = (absoluteUrl: string): string => {
  let scheme = 'http'
  let rest: string
  try {
    const parsed = new URL(absoluteUrl)
    scheme = parsed.protocol.replace(/:$/, '') || scheme
    rest = absoluteUrl.slice(parsed.protocol.length)
  } catch {
    rest = absoluteUrl.replace(/^[a-z][a-z0-9+.-]*:/i, '')
  }
  return `intent:${rest}#Intent;scheme=${scheme};type=video/*;end`
}

/** ponytail: Infuse scheme length cap. Longer season lists fall back to one m3u url. */
export const INFUSE_PLAYLIST_HREF_LIMIT = 8000

/** Deep-link scheme for a single enabled external player and an absolute stream URL. */
export const buildExternalPlayerHref = (
  player: ExternalPosterPlayAction,
  fullLink: string,
  platform: PlayerLinkPlatform = {},
): string => {
  const encoded = encodeURIComponent(fullLink)
  switch (player) {
    case 'infuse':
      return `infuse://x-callback-url/play?url=${encoded}`
    case 'senPlayer':
      return `senplayer://x-callback-url/play?url=${encoded}`
    case 'vlc':
      if (platform.ios) return `vlc-x-callback://x-callback-url/stream?url=${encoded}`
      if (platform.android) return buildAndroidIntentHref(fullLink)
      return `vlc://${fullLink}`
    case 'iina':
      return `iina://weblink?url=${encoded}`
    default:
      return fullLink
  }
}

/**
 * Infuse plays every `url` in order as a temporary playlist.
 * Over {@link INFUSE_PLAYLIST_HREF_LIMIT} characters, one m3u `url` instead.
 */
export const buildInfusePlaylistHref = (episodes: { url: string; filename: string }[], m3uFallback: string): string => {
  const href =
    'infuse://x-callback-url/play?' +
    episodes
      .map(episode => `url=${encodeURIComponent(episode.url)}&filename=${encodeURIComponent(episode.filename)}`)
      .join('&')
  if (href.length > INFUSE_PLAYLIST_HREF_LIMIT) return buildExternalPlayerHref('infuse', m3uFallback)
  return href
}

/** Player deep link for the Playlist button. Download and copy stay on the page. */
export const playlistActionHref = (
  action: PlaylistButtonAction,
  absoluteUrl: string,
  platform: PlayerLinkPlatform = {},
): string | null => {
  if (action === 'download' || action === 'copyLink') return null
  return buildExternalPlayerHref(action, absoluteUrl, platform)
}
