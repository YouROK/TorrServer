import { Modal, useOverlayState } from '@heroui/react'
import { useEffect, type CSSProperties, type ReactNode } from 'react'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'
import { DIALOG_FULLSCREEN } from 'shared/ui/dialogSizes'

export interface AppDialogProps {
  open: boolean
  onClose: () => void
  children: ReactNode
  size?: 'sm' | 'md' | 'lg' | 'full'
  fullScreen?: boolean
  className?: string
  /** Extra classes applied to the dialog surface itself — use to widen a dialog past its `size` ceiling. */
  dialogClassName?: string
  /**
   * Inline min/max-width (and optional fixed height) for the dialog surface.
   * Ignored when `fullScreen` — then `DIALOG_FULLSCREEN` (`100dvh`) is applied instead.
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
  fullScreen = false,
  className,
  dialogClassName,
  dialogStyle,
}: AppDialogProps) {
  useSyncModalOpen(open)

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
        <Modal.Container size={fullScreen ? 'full' : size} scroll='inside' className={className}>
          <Modal.Dialog className={dialogClassName} style={fullScreen ? DIALOG_FULLSCREEN : dialogStyle}>
            {children}
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal.Root>
  )
}
