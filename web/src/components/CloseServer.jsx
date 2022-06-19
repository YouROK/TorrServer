import { useState } from 'react'
import { Button, Dialog, DialogActions, DialogTitle, ListItemIcon, ListItemText } from '@material-ui/core'
import { PowerSettingsNew as PowerSettingsNewIcon } from '@material-ui/icons'
import { shutdownHost } from 'utils/Hosts'
import { useTranslation } from 'react-i18next'
import { isStandaloneApp } from 'utils/Utils'
import StyledMenuButtonWrapper from 'style/StyledMenuButtonWrapper'

export default function CloseServer({ isOffline, isLoading }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const closeDialog = () => setOpen(false)
  const openDialog = () => setOpen(true)

  return (
    <>
      <StyledMenuButtonWrapper disabled={isOffline || isLoading} button key={t('CloseServer')} onClick={openDialog}>
        {isStandaloneApp ? (
          <>
            <PowerSettingsNewIcon />
            <div>{t('CloseServer')}</div>
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

      <Dialog open={open} onClose={closeDialog}>
        <DialogTitle>{t('CloseServer?')}</DialogTitle>
        <DialogActions>
          <Button variant='outlined' onClick={closeDialog} color='secondary'>
            {t('Cancel')}
          </Button>

          <Button
            variant='contained'
            onClick={() => {
              fetch(shutdownHost())
              closeDialog()
            }}
            color='secondary'
            autoFocus
          >
            {t('TurnOff')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}
