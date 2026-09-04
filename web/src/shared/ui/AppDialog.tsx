import { Modal, useOverlayState } from '@heroui/react'
import { useEffect, type CSSProperties, type ReactNode } from 'react'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'
import { DIALOG_FULLSCREEN } from 'shared/ui/dialogSizes'

export interface AppDialogProps {
  open: boolean
  onClose: () => void
  children: ReactNode
  size?: 'sm' | 'md' | 'lg' | 'full'
  /**
   * Override fullscreen policy. Omit to follow {@link useDialogFullScreen}
   * (phone / tablet landscape / Appearance force). Pass `false` only together
   * with `compact` for short confirms that stay centered cards.
   */
  fullScreen?: boolean
  /**
   * Short confirm / tip sheets — stay a centered card even when the global
   * fullscreen policy (or ≤960 CSS) would otherwise stretch them edge-to-edge.
   * Adds `.ts-compact-modal` and forces windowed geometry.
   */
  compact?: boolean
  className?: string
  /** Extra classes applied to the dialog surface itself — use to widen a dialog past its `size` ceiling. */
  dialogClassName?: string
  /**
   * Inline min/max-width (and optional fixed height) for the dialog surface.
   * Ignored when fullscreen — then `DIALOG_FULLSCREEN` (`height: 100%` of the
   * remapped fullscreen container) is applied instead.
   * HeroUI's `size` ceilings and our collapse-prevention floor (`index.css`) live in CSS
   * layers, so a plain Tailwind width utility can lose to them regardless of specificity —
   * inline style always wins and is the only reliable way to widen a dialog past `lg`.
   */
  dialogStyle?: CSSProperties
}

/** Modal wrapper that registers open state for bottom-nav / chrome coordination. */
export default function AppDialog({
  open,
  onClose,
  children,
  size = 'md',
  fullScreen,
  compact = false,
  className,
  dialogClassName,
  dialogStyle,
}: AppDialogProps) {
  useSyncModalOpen(open)
  const policyFullScreen = useDialogFullScreen()
  const isFullScreen = compact ? false : (fullScreen ?? policyFullScreen)
  const dialogClass =
    [dialogClassName, compact ? 'ts-compact-modal' : null, isFullScreen ? 'modal__dialog--full' : null]
      .filter(Boolean)
      .join(' ') || undefined
  const containerClass =
    [className, isFullScreen ? 'modal__container--full' : null].filter(Boolean).join(' ') || undefined

  const state = useOverlayState({
    isOpen: open,
    onOpenChange: next => {
      if (!next) onClose()
    },
  })

  useEffect(() => {
    if (open) state.setOpen(true)
  }, [open, state])

  return (
    <Modal.Root state={state}>
      <Modal.Backdrop>
        <Modal.Container
          size={isFullScreen ? 'full' : size}
          scroll='inside'
          placement={isFullScreen ? 'center' : 'auto'}
          className={containerClass}
        >
          <Modal.Dialog className={dialogClass} style={isFullScreen ? DIALOG_FULLSCREEN : dialogStyle}>
            {children}
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal.Root>
  )
}
