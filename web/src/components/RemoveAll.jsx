import { Button, Dialog, DialogActions, DialogTitle } from '@material-ui/core'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import DeleteIcon from '@material-ui/icons/Delete'
import { useState } from 'react'
import { torrentsHost } from 'utils/Hosts'
import { useTranslation } from 'react-i18next'
import { ThemeProvider } from '@material-ui/core/styles'
import { lightTheme } from 'components/App'

const fnRemoveAll = () => {
  fetch(torrentsHost(), {
    method: 'post',
    body: JSON.stringify({ action: 'list' }),
    headers: {
      Accept: 'application/json, text/plain, */*',
      'Content-Type': 'application/json',
    },
  })
    .then(res => res.json())
    .then(json => {
      json.forEach(torr => {
        fetch(torrentsHost(), {
          method: 'post',
          body: JSON.stringify({ action: 'rem', hash: torr.hash }),
          headers: {
            Accept: 'application/json, text/plain, */*',
            'Content-Type': 'application/json',
          },
        })
      })
    })
}

export default function RemoveAll() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const closeDialog = () => setOpen(false)
  const openDialog = () => setOpen(true)
  return (
    <>
      <ListItem button key={t('RemoveAll')} onClick={openDialog}>
        <ListItemIcon>
          <DeleteIcon />
        </ListItemIcon>

        <ListItemText primary={t('RemoveAll')} />
      </ListItem>

      <ThemeProvider theme={lightTheme}>
        <Dialog open={open} onClose={closeDialog}>
          <DialogTitle>{t('DeleteTorrents?')}</DialogTitle>
          <DialogActions>
            <Button variant='outlined' onClick={closeDialog} color='primary'>
              {t('Cancel')}
            </Button>

            <Button
              variant='contained'
              onClick={() => {
                fnRemoveAll()
                closeDialog()
              }}
              color='primary'
              autoFocus
            >
              {t('OK')}
            </Button>
          </DialogActions>
        </Dialog>
      </ThemeProvider>
    </>
  )
}
