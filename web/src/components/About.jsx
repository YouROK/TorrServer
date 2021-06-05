import axios from 'axios'
import { useEffect, useState } from 'react'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import InfoIcon from '@material-ui/icons/Info'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useTranslation } from 'react-i18next'
import { echoHost } from 'utils/Hosts'

export default function AboutDialog() {
  const [open, setOpen] = useState(false)
  // eslint-disable-next-line no-unused-vars
  const { t } = useTranslation()
  const [torrServerVersion, setTorrServerVersion] = useState('')
  useEffect(() => {
    axios.get(echoHost()).then(({ data }) => setTorrServerVersion(data))
  }, [])

  return (
    <div>
      <ListItem button key='Settings' onClick={() => setOpen(true)}>
        <ListItemIcon>
          <InfoIcon />
        </ListItemIcon>
        <ListItemText primary={t('About')} />
      </ListItem>

      <Dialog open={open} onClose={() => setOpen(false)} aria-labelledby='form-dialog-title' fullWidth maxWidth='lg'>
        <DialogTitle id='form-dialog-title'>{t('About')}</DialogTitle>

        <DialogContent>
          <center>
            <h2>TorrServer {torrServerVersion}</h2>
            <a href='https://github.com/YouROK/TorrServer'>https://github.com/YouROK/TorrServer</a>
          </center>
          <DialogContent>
            <center>
              <h2>{t('ThanksToEveryone')}</h2>
            </center>
            <br />
            <h2>{t('SpecialThanks')}</h2>
            <b>anacrolix Matt Joiner</b> <a href='https://github.com/anacrolix/'>github.com/anacrolix</a>
            <br />
            <b>nikk</b> <a href='https://github.com/tsynik'>github.com/tsynik</a>
            <br />
            <b>dancheskus</b> <a href='https://github.com/dancheskus'>github.com/dancheskus</a>
            <br />
            <b>tw1cker Руслан Пахнев</b> <a href='https://github.com/Nemiroff'>github.com/Nemiroff</a>
            <br />
            <b>SpAwN_LMG</b>
            <br />
          </DialogContent>
        </DialogContent>

        <DialogActions>
          <Button onClick={() => setOpen(false)} color='primary' variant='outlined' autoFocus>
            {t('Close')}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}
