import { Button, Modal } from '@heroui/react'
import { Power } from 'lucide-react'
import { iconAction, iconMenu } from 'shared/ui/iconProps'
import { useTranslation } from 'react-i18next'

import { shutdownHost } from 'shared/api/hosts'
import { authFetch } from 'shared/api/authCredentials'
import AppDialog from 'shared/ui/AppDialog'
import UnsafeButton from 'shared/ui/UnsafeButton'

export interface CloseServerDialogProps {
  open: boolean
  onClose: () => void
}

/** Destructive confirm dialog for shutting down the TorrServer process. */
export default function CloseServerDialog({ open, onClose }: CloseServerDialogProps) {
  const { t } = useTranslation()

  return (
    <AppDialog open={open} onClose={onClose} size='sm' compact>
      <Modal.Header className='shrink-0'>
        <Modal.Icon className='bg-danger/15 text-danger'>
          <Power {...iconAction} aria-hidden />
        </Modal.Icon>
        <Modal.Heading>{t('CloseServer?')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body>{t('ConfirmCloseServer')}</Modal.Body>
      <Modal.Footer className='shrink-0'>
        <Button variant='secondary' onPress={onClose} autoFocus>
          {t('Cancel')}
        </Button>
        <UnsafeButton
          variant='danger'
          onPress={() => {
            void authFetch(shutdownHost())
            onClose()
          }}
        >
          <Power {...iconMenu} aria-hidden />
          {t('TurnOff')}
        </UnsafeButton>
      </Modal.Footer>
    </AppDialog>
  )
}
