import { useState } from 'react'
import { Button, Dialog, DialogActions, DialogTitle, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'
import { PowerSettingsNew as PowerSettingsNewIcon } from '@material-ui/icons'
import { shutdownHost } from 'utils/Hosts'
import { useTranslation } from 'react-i18next'
import { ThemeProvider } from '@material-ui/core/styles'
import { lightTheme } from 'components/App'

export default function CloseServer() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const closeDialog = () => setOpen(false)
  const openDialog = () => setOpen(true)

  return (
    <>
      <ListItem button key={t('CloseServer')} onClick={openDialog}>
        <ListItemIcon>
          <PowerSettingsNewIcon />
        </ListItemIcon>

        <ListItemText primary={t('CloseServer')} />
      </ListItem>

      <ThemeProvider theme={lightTheme}>
        <Dialog open={open} onClose={closeDialog}>
          <DialogTitle>{t('CloseServer?')}</DialogTitle>
          <DialogActions>
            <Button variant='outlined' onClick={closeDialog} color='primary'>
              {t('Cancel')}
            </Button>

            <Button
              variant='contained'
              onClick={() => {
                fetch(shutdownHost())
                closeDialog()
              }}
              color='primary'
              autoFocus
            >
              {t('TurnOff')}
            </Button>
          </DialogActions>
        </Dialog>
      </ThemeProvider>
    </>
  )
}
