import { useState } from 'react'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogTitle from '@material-ui/core/DialogTitle'
import InfoIcon from '@material-ui/icons/Info'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'

export default function AboutDialog() {
  const [open, setOpen] = useState(false)

  return (
    <div>
      <ListItem button key='Settings' onClick={() => setOpen(true)}>
        <ListItemIcon>
          <InfoIcon />
        </ListItemIcon>
        <ListItemText primary='About' />
      </ListItem>

      <Dialog open={open} onClose={() => setOpen(false)} aria-labelledby='form-dialog-title' fullWidth maxWidth='lg'>
        <DialogTitle id='form-dialog-title'>About</DialogTitle>

        <DialogContent>
          <DialogContent>
            <DialogContentText id='alert-dialog-description'>
              <center>
                <h2>Thanks to everyone who tested and helped.</h2>
              </center>
              <br />
              <h2>Special thanks:</h2>
              <b>Anacrolix Matt Joiner</b> <a href='https://github.com/anacrolix/'>github.com/anacrolix</a>
              <br />
              <b>tsynik nikk Никита</b> <a href='https://github.com/tsynik'>github.com/tsynik</a>
              <br />
              <b>dancheskus</b> <a href='https://github.com/dancheskus'>github.com/dancheskus</a>
              <br />
              <b>Tw1cker Руслан Пахнев</b> <a href='https://github.com/Nemiroff'>github.com/Nemiroff</a>
              <br />
              <b>SpAwN_LMG</b>
              <br />
            </DialogContentText>
          </DialogContent>
        </DialogContent>

        <DialogActions>
          <Button onClick={() => setOpen(false)} color='primary' variant='outlined' autoFocus>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}
