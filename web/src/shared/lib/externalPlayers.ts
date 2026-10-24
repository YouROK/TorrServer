import { useTranslation } from 'react-i18next'
import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import { detectStandaloneApp, isAppleDevice, isMacOS, playerLinkPlatform } from 'shared/lib/platform'
import type { PosterPlayPlayerFlags } from 'shared/lib/posterPlay'
import { buildExternalPlayerHref, buildInfusePlaylistHref } from 'shared/lib/streamLinks'

export {
  buildExternalPlayerHref,
  buildInfusePlaylistHref,
  INFUSE_PLAYLIST_HREF_LIMIT,
  playlistActionHref,
} from 'shared/lib/streamLinks'

export interface ExternalPlayerLink {
  label: string
  href: string
}

/**
 * Reads the user's configured mobile/desktop player preferences (Settings → App) and exposes a
 * builder for the matching deep-link scheme per direct-stream URL. Shared between the per-file
 * action row and the torrent card's quick-open action so both stay in sync.
 */
export function useExternalPlayers() {
  const { t } = useTranslation()
  const [isVlcUsed] = useLocalBoolPref('isVlcUsed')
  const [isInfuseUsed] = useLocalBoolPref('isInfuseUsed')
  const [isSenPlayerUsed] = useLocalBoolPref('isSenPlayerUsed')
  const [isIinaUsed] = useLocalBoolPref('isIinaUsed')
  const isStandaloneApp = detectStandaloneApp()
  const isMac = isMacOS()
  const isApple = isAppleDevice()
  const playerPlatform = playerLinkPlatform()

  const playerFlags: PosterPlayPlayerFlags = {
    isVlcUsed,
    isInfuseUsed,
    isSenPlayerUsed,
    isIinaUsed,
    isApple,
    isMac,
  }

  const hasAnyExternalPlayer =
    (isApple && isInfuseUsed) || (isApple && isSenPlayerUsed) || isVlcUsed || (isMac && isIinaUsed)

  /** Whether the plain "open direct link" fallback should also be shown alongside app deep links. */
  const shouldShowOpenLink = !isStandaloneApp || !hasAnyExternalPlayer

  const buildExternalPlayers = (fullLink: string): ExternalPlayerLink[] => {
    const links: ExternalPlayerLink[] = []
    if (isApple && isInfuseUsed) links.push({ label: t('Infuse'), href: buildExternalPlayerHref('infuse', fullLink) })
    if (isApple && isSenPlayerUsed)
      links.push({ label: t('SenPlayer'), href: buildExternalPlayerHref('senPlayer', fullLink) })
    if (isVlcUsed) links.push({ label: t('VLC'), href: buildExternalPlayerHref('vlc', fullLink, playerPlatform) })
    if (isMac && isIinaUsed) links.push({ label: t('IINA'), href: buildExternalPlayerHref('iina', fullLink) })
    return links
  }

  /** Series footer: Infuse gets one url per episode; the others open the torrent m3u. */
  const buildSeriesExternalPlayers = (
    playlistUrl: string,
    episodes: { url: string; filename: string }[],
  ): ExternalPlayerLink[] => {
    const links: ExternalPlayerLink[] = []
    if (isApple && isInfuseUsed)
      links.push({ label: t('Infuse'), href: buildInfusePlaylistHref(episodes, playlistUrl) })
    if (isApple && isSenPlayerUsed)
      links.push({ label: t('SenPlayer'), href: buildExternalPlayerHref('senPlayer', playlistUrl) })
    if (isVlcUsed) links.push({ label: t('VLC'), href: buildExternalPlayerHref('vlc', playlistUrl, playerPlatform) })
    if (isMac && isIinaUsed) links.push({ label: t('IINA'), href: buildExternalPlayerHref('iina', playlistUrl) })
    return links
  }

  return {
    buildExternalPlayers,
    buildSeriesExternalPlayers,
    shouldShowOpenLink,
    hasAnyExternalPlayer,
    playerFlags,
    playerPlatform,
  }
}
