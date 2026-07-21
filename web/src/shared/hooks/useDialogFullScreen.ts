import { useMediaQuery } from '@heroui/react'

import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import { MEDIA_TABLET_LANDSCAPE, queryMax } from 'shared/theme/breakpoints'

/** localStorage key — also used by Appearance settings switch. */
export const DIALOGS_FULLSCREEN_PREF = 'dialogsFullScreen'

/**
 * Sheet dialogs go fullscreen when:
 * - viewport ≤ dialog breakpoint (960px), or
 * - tablet landscape (iPad etc.), or
 * - Appearance → Dialogs fullscreen is on.
 *
 * VideoPlayer keeps its own media query.
 */
export function useDialogFullScreen(): boolean {
  const narrow = useMediaQuery(queryMax('dialog'))
  const tabletLandscape = useMediaQuery(MEDIA_TABLET_LANDSCAPE)
  const [force] = useLocalBoolPref(DIALOGS_FULLSCREEN_PREF, false)
  return narrow || tabletLandscape || force
}
