import { useTranslation } from 'react-i18next'
import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import { detectStandaloneApp, isAppleDevice, isMacOS } from 'shared/lib/platform'
import type { ExternalPosterPlayAction, PosterPlayPlayerFlags } from 'shared/lib/posterPlay'

export interface ExternalPlayerLink {
  label: string
  href: string
}

/** Deep-link scheme for a single enabled external player and an absolute stream URL. */
export const buildExternalPlayerHref = (player: ExternalPosterPlayAction, fullLink: string): string => {
  const encoded = encodeURIComponent(fullLink)
  switch (player) {
    case 'infuse':
      return `infuse://x-callback-url/play?url=${encoded}`
    case 'senPlayer':
      return `senplayer://x-callback-url/play?url=${encoded}`
    case 'vlc':
      return `vlc://${fullLink}`
    case 'iina':
      return `iina://weblink?url=${encoded}`
    default:
      return fullLink
  }
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
    if (isVlcUsed) links.push({ label: t('VLC'), href: buildExternalPlayerHref('vlc', fullLink) })
    if (isMac && isIinaUsed) links.push({ label: t('IINA'), href: buildExternalPlayerHref('iina', fullLink) })
    return links
  }

  return { buildExternalPlayers, shouldShowOpenLink, hasAnyExternalPlayer, playerFlags }
}
