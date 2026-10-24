import { useLocalJsonPref } from 'shared/hooks/useLocalPref'
import { playlistTorrHost, streamHost } from 'shared/api/hosts'

/** localStorage key for poster Play button behavior (Settings → Players). */
export const POSTER_PLAY_ACTION_KEY = 'posterPlayAction'

/** localStorage key for the Playlist button (Settings → Players). */
export const PLAYLIST_BUTTON_ACTION_KEY = 'playlistButtonAction'

export type PosterPlayAction = 'builtin' | 'copyLink' | 'vlc' | 'infuse' | 'senPlayer' | 'iina'

/** `download` keeps the m3u link. A player opens that m3u in the app. */
export type PlaylistButtonAction = 'download' | 'copyLink' | ExternalPosterPlayAction

export type ExternalPosterPlayAction = Exclude<PosterPlayAction, 'builtin' | 'copyLink'>

export interface PosterPlayPlayerFlags {
  isVlcUsed: boolean
  isInfuseUsed: boolean
  isSenPlayerUsed: boolean
  isIinaUsed: boolean
  isApple: boolean
  isMac: boolean
}

const ALL_ACTIONS: readonly PosterPlayAction[] = ['builtin', 'copyLink', 'vlc', 'infuse', 'senPlayer', 'iina'] as const

export const isPosterPlayAction = (value: unknown): value is PosterPlayAction =>
  typeof value === 'string' && (ALL_ACTIONS as readonly string[]).includes(value)

/** Platform default: copy stream/playlist link on iPhone/iPad; built-in player elsewhere. */
export const defaultPosterPlayAction = (isIOS: boolean): PosterPlayAction => (isIOS ? 'copyLink' : 'builtin')

/** Whether an external player action is currently enabled for this platform. */
export const isExternalPosterPlayEnabled = (
  action: ExternalPosterPlayAction,
  flags: PosterPlayPlayerFlags,
): boolean => {
  switch (action) {
    case 'vlc':
      return flags.isVlcUsed
    case 'infuse':
      return flags.isApple && flags.isInfuseUsed
    case 'senPlayer':
      return flags.isApple && flags.isSenPlayerUsed
    case 'iina':
      return flags.isMac && flags.isIinaUsed
    default:
      return false
  }
}

/**
 * Coerce a stored pref to a usable action. Disabled / platform-unavailable externals
 * fall back to copyLink on iOS and builtin elsewhere.
 */
export const coercePosterPlayAction = (
  stored: unknown,
  flags: PosterPlayPlayerFlags,
  isIOS: boolean,
): PosterPlayAction => {
  const fallback = defaultPosterPlayAction(isIOS)
  if (!isPosterPlayAction(stored)) return fallback
  if (stored === 'builtin' || stored === 'copyLink') return stored
  return isExternalPosterPlayEnabled(stored, flags) ? stored : fallback
}

const PLAYLIST_ACTIONS = ['download', 'copyLink', 'vlc', 'infuse', 'senPlayer', 'iina'] as const

export const isPlaylistButtonAction = (value: unknown): value is PlaylistButtonAction =>
  typeof value === 'string' && (PLAYLIST_ACTIONS as readonly string[]).includes(value)

/** i18n key for a playlist-button option or the player name on the button. */
export const playlistButtonLabelKey = (action: PlaylistButtonAction): string => {
  switch (action) {
    case 'download':
      return 'DownloadPlaylist'
    case 'copyLink':
      return 'CopyLink'
    case 'vlc':
      return 'VLC'
    case 'infuse':
      return 'Infuse'
    case 'senPlayer':
      return 'SenPlayer'
    case 'iina':
      return 'IINA'
    default:
      return 'DownloadPlaylist'
  }
}

/** Disabled players fall back to download so the button stays a real playlist link. */
export const coercePlaylistButtonAction = (stored: unknown, flags: PosterPlayPlayerFlags): PlaylistButtonAction => {
  if (!isPlaylistButtonAction(stored)) return 'download'
  if (stored === 'download' || stored === 'copyLink') return stored
  return isExternalPosterPlayEnabled(stored, flags) ? stored : 'download'
}

/** Settings → Players value for the Playlist button, coerced when a player toggle is off. */
export function usePlaylistButtonAction(flags: PosterPlayPlayerFlags): PlaylistButtonAction {
  const [stored] = useLocalJsonPref<unknown>(PLAYLIST_BUTTON_ACTION_KEY, 'download')
  return coercePlaylistButtonAction(stored, flags)
}

/** Download and copy, plus external players that are enabled for this platform. */
export const availablePlaylistButtonActions = (flags: PosterPlayPlayerFlags): PlaylistButtonAction[] => {
  const actions: PlaylistButtonAction[] = ['download', 'copyLink']
  if (flags.isVlcUsed) actions.push('vlc')
  if (flags.isApple && flags.isInfuseUsed) actions.push('infuse')
  if (flags.isApple && flags.isSenPlayerUsed) actions.push('senPlayer')
  if (flags.isMac && flags.isIinaUsed) actions.push('iina')
  return actions
}

/** Options shown in Settings for the current platform + enabled player toggles. */
export const availablePosterPlayActions = (flags: PosterPlayPlayerFlags): PosterPlayAction[] => {
  const actions: PosterPlayAction[] = ['builtin', 'copyLink']
  if (flags.isVlcUsed) actions.push('vlc')
  if (flags.isApple && flags.isInfuseUsed) actions.push('infuse')
  if (flags.isApple && flags.isSenPlayerUsed) actions.push('senPlayer')
  if (flags.isMac && flags.isIinaUsed) actions.push('iina')
  return actions
}

export const magnetFromHash = (hash: string, displayName: string): string =>
  `magnet:?xt=urn:btih:${hash}&dn=${encodeURIComponent(displayName)}`

const fileBaseName = (path: string): string => path.split('\\').pop()?.split('/').pop() || path

/** Query string for `/stream`. Flag keys (`play`, `m3u`) stay valueless — Gin `GetQuery` only checks presence. */
const streamQuery = (hash: string, params: Record<string, string>, flags: readonly string[]): string => {
  const query = new URLSearchParams({ link: hash, ...params }).toString()
  return flags.length ? `${query}&${flags.join('&')}` : query
}

/** Direct play stream URL. Path segment and `link` are encoded (`#`, `?`, `&`, `%`, non-ASCII). */
export const streamPlayUrl = (hash: string, file: { id: number; path: string }): string =>
  `${streamHost()}/${encodeURIComponent(fileBaseName(file.path))}?${streamQuery(hash, { index: String(file.id) }, ['play'])}`

/** Buffer a file without opening it (`GET /stream?…&preload`). */
export const streamPreloadUrl = (hash: string, fileId: number): string =>
  `${streamHost()}?${streamQuery(hash, { index: String(fileId) }, ['preload'])}`

/**
 * M3U playlist URL for the whole torrent.
 * `fn=file.m3u` matches server `all.m3u` entries — ForkPlayer mishandles a link whose path ends in `.m3u`.
 */
export const torrentPlaylistUrl = (hash: string, displayName: string): string =>
  `${playlistTorrHost()}/${encodeURIComponent(displayName)}.m3u?${streamQuery(hash, {}, ['m3u', 'fn=file.m3u'])}`
