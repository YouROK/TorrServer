import { useState } from 'react'
import { Button, DialogActions, DialogTitle, ListItemIcon, ListItemText } from '@material-ui/core'
import { StyledDialog, StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import { PowerSettingsNew as PowerSettingsNewIcon, PowerOff as PowerOffIcon } from '@material-ui/icons'
import { shutdownHost } from 'utils/Hosts'
import { useTranslation } from 'react-i18next'
import { isStandaloneApp } from 'utils/Utils'
import useOnStandaloneAppOutsideClick from 'utils/useOnStandaloneAppOutsideClick'

import UnsafeButton from './UnsafeButton'

export default function CloseServer({ isOffline, isLoading }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const closeDialog = () => setOpen(false)
  const openDialog = () => setOpen(true)

  const ref = useOnStandaloneAppOutsideClick(closeDialog)

  return (
    <>
      <StyledMenuButtonWrapper disabled={isOffline || isLoading} button key={t('CloseServer')} onClick={openDialog}>
        {isStandaloneApp ? (
          <>
            <PowerSettingsNewIcon />
            <div>{t('TurnOff')}</div>
          </>
        ) : (
          <>
            <ListItemIcon>
              <PowerSettingsNewIcon />
            </ListItemIcon>

            <ListItemText primary={t('CloseServer')} />
          </>
        )}
      </StyledMenuButtonWrapper>

      <StyledDialog open={open} onClose={closeDialog} ref={ref}>
        <DialogTitle>{t('CloseServer?')}</DialogTitle>
        <DialogActions>
          <Button variant='outlined' onClick={closeDialog} color='secondary'>
            {t('Cancel')}
          </Button>

          <UnsafeButton
            timeout={5}
            startIcon={<PowerOffIcon />}
            variant='contained'
            onClick={() => {
              fetch(shutdownHost())
              closeDialog()
            }}
            color='secondary'
            autoFocus
          >
            {t('TurnOff')}
          </UnsafeButton>
        </DialogActions>
      </StyledDialog>
    </>
  )
}
