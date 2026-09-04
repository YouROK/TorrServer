import { useEffect } from 'react'
import { useMediaQuery } from '@heroui/react'
import type { CSSProperties } from 'react'

import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import { MEDIA_TABLET_LANDSCAPE, queryMax } from 'shared/theme/breakpoints'
import { DIALOG_FULLSCREEN } from 'shared/ui/dialogSizes'

/** localStorage key — also used by Appearance settings switch. */
export const DIALOGS_FULLSCREEN_PREF = 'dialogsFullScreen'

/** Pure policy used by {@link useDialogFullScreen} (and unit tests). */
export function resolveDialogFullScreen(input: { narrow: boolean; tabletLandscape: boolean; force: boolean }): boolean {
  return input.narrow || input.tabletLandscape || input.force
}

/**
 * Sheet dialogs go fullscreen when:
 * - viewport ≤ dialog breakpoint (960px), or
 * - tablet landscape (touch iPad-class devices), or
 * - Appearance → Dialogs fullscreen is on (forces wide monitors too).
 *
 * Mirrors policy onto `html[data-dialogs-fullscreen]` for CSS above 960px.
 * VideoPlayer and raw HeroUI modals should use the same hook / {@link useDialogFullLayout}.
 */
export function useDialogFullScreen(): boolean {
  const narrow = useMediaQuery(queryMax('dialog'))
  const tabletLandscape = useMediaQuery(MEDIA_TABLET_LANDSCAPE)
  const [force] = useLocalBoolPref(DIALOGS_FULLSCREEN_PREF, false)
  const full = resolveDialogFullScreen({ narrow, tabletLandscape, force })

  useEffect(() => {
    const root = document.documentElement
    if (full) root.dataset.dialogsFullscreen = '1'
    else delete root.dataset.dialogsFullscreen
  }, [full])

  return full
}

type DialogSize = 'sm' | 'md' | 'lg' | 'full'

/** Size / placement / inline style for HeroUI Modal.Container + Modal.Dialog. */
export function useDialogFullLayout(base: { size?: Exclude<DialogSize, 'full'>; dialogStyle?: CSSProperties } = {}) {
  const full = useDialogFullScreen()
  const size: DialogSize = full ? 'full' : (base.size ?? 'md')
  return {
    full,
    size,
    placement: (full ? 'center' : 'auto') as 'center' | 'auto',
    dialogStyle: full ? DIALOG_FULLSCREEN : base.dialogStyle,
  }
}
