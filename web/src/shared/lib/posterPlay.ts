import { playlistTorrHost, streamHost } from 'shared/api/hosts'

/** localStorage key for poster Play button behavior (Settings → Players). */
export const POSTER_PLAY_ACTION_KEY = 'posterPlayAction'

export type PosterPlayAction = 'builtin' | 'copyLink' | 'vlc' | 'infuse' | 'senPlayer' | 'iina'

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

/** Direct play stream URL (relative to the current origin when used with `new URL(..., location)`). */
export const streamPlayUrl = (hash: string, file: { id: number; path: string }): string =>
  `${streamHost()}/${encodeURIComponent(fileBaseName(file.path))}?link=${hash}&index=${file.id}&play`

/** M3U playlist URL for the whole torrent (same shape as Download playlist). */
export const torrentPlaylistUrl = (hash: string, displayName: string): string =>
  `${playlistTorrHost()}/${encodeURIComponent(displayName)}.m3u?link=${hash}&m3u`
