import { useState } from 'react'
import { Button, Dialog, DialogActions, DialogTitle, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'
import { PowerSettingsNew as PowerSettingsNewIcon } from '@material-ui/icons'
import { shutdownHost } from 'utils/Hosts'

export default function CloseServer() {
  const [open, setOpen] = useState(false)
  const closeDialog = () => setOpen(false)
  const openDialog = () => setOpen(true)

  return (
    <>
      <ListItem button key='Close server' onClick={openDialog}>
        <ListItemIcon>
          <PowerSettingsNewIcon />
        </ListItemIcon>

        <ListItemText primary='Close server' />
      </ListItem>

      <Dialog open={open} onClose={closeDialog}>
        <DialogTitle>Close server?</DialogTitle>
        <DialogActions>
          <Button variant='outlined' onClick={closeDialog} color='primary'>
            Cancel
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
            Ok
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}
