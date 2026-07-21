import { Button, Link, Modal } from '@heroui/react'
import { Heart, SquareArrowOutUpRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import { iconAction } from 'shared/ui/iconProps'

import { DONATE_LINKS } from './donateLinks'

export interface DonateDialogProps {
  open: boolean
  onClose: () => void
}

/** Support dialog: Boosty / YooMoney / TBank. */
export default function DonateDialog({ open, onClose }: DonateDialogProps) {
  const { t } = useTranslation()
  const isFullScreenBreakpoint = useDialogFullScreen()

  return (
    <AppDialog open={open} onClose={onClose} size='sm' fullScreen={isFullScreenBreakpoint}>
      <Modal.Header>
        <Modal.Icon className='bg-accent/15 text-accent'>
          <Heart {...iconAction} aria-hidden />
        </Modal.Icon>
        <Modal.Heading>{t('Donate')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body>
        <ul className='flex flex-col gap-2 sm:flex-row sm:flex-wrap'>
          {DONATE_LINKS.map(link => (
            <li key={link.id} className='w-full sm:min-w-[8.5rem] sm:flex-1'>
              <Link
                href={link.href}
                target='_blank'
                rel='noopener noreferrer'
                className='flex min-h-11 w-full items-center justify-between gap-2 rounded-lg border border-border bg-surface-secondary px-3 py-2.5 text-sm font-medium text-foreground transition-colors hover-fine:border-accent/40 hover-fine:bg-accent-soft/40'
              >
                <span>{link.name}</span>
                <SquareArrowOutUpRight size={14} strokeWidth={1.75} className='shrink-0 text-muted' aria-hidden />
              </Link>
            </li>
          ))}
        </ul>
      </Modal.Body>
      <Modal.Footer>
        <Button variant='secondary' onPress={onClose} autoFocus>
          {t('Close')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}
