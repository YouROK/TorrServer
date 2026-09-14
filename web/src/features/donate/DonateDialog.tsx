import { Link, Modal } from '@heroui/react'
import { Heart, SquareArrowOutUpRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import AppDialog from 'shared/ui/AppDialog'
import { iconAction } from 'shared/ui/iconProps'

import { DONATE_LINKS } from './donateLinks'
import { DONATE_LINKS_PP } from './donateLinks'
import { DONATE_LINKS_NG } from './donateLinks'

export interface DonateDialogProps {
  open: boolean
  onClose: () => void
}

/** Support dialog: Boosty / YooMoney / TBank. */
export default function DonateDialog({ open, onClose }: DonateDialogProps) {
  const { t } = useTranslation()

  return (
    <AppDialog open={open} onClose={onClose} size='sm' compact>
      <Modal.Header className='shrink-0'>
        <Modal.Icon className='bg-accent/15 text-accent'>
          <Heart {...iconAction} aria-hidden />
        </Modal.Icon>
        <Modal.Heading>{t('Donate')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body>
        <h2 className='font-semibold mb-2' title='YouROK'>
          @YouROK
        </h2>
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
        <h2 className='font-semibold mt-3 mb-2' title='nikk'>
          @tsynik
        </h2>
        <ul className='flex flex-col gap-2 sm:flex-row sm:flex-wrap'>
          {DONATE_LINKS_NG.map(link => (
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
        <h2 className='font-semibold mt-3 mb-2' title='Pavel Pikta'>
          @pavelpikta
        </h2>
        <ul className='flex flex-col gap-2 sm:flex-row sm:flex-wrap'>
          {DONATE_LINKS_PP.map(link => (
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
    </AppDialog>
  )
}
