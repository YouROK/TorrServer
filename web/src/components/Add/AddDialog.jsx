import { useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import { torrentsHost } from 'utils/Hosts'
import axios from 'axios'
import { useTranslation } from 'react-i18next'

export default function AddDialog({ handleClose }) {
  const [link, setLink] = useState('')
  const [title, setTitle] = useState('')
  const [poster, setPoster] = useState('')

  const inputMagnet = ({ target: { value } }) => setLink(value)
  const inputTitle = ({ target: { value } }) => setTitle(value)
  const inputPoster = ({ target: { value } }) => setPoster(value)

  // eslint-disable-next-line no-unused-vars
  const { t } = useTranslation()

  const handleSave = () => {
    axios.post(torrentsHost(), { action: 'add', link, title, poster, save_to_db: true }).finally(() => handleClose())
  }

  return (
    <Dialog open onClose={handleClose} aria-labelledby='form-dialog-title' fullWidth>
      <DialogTitle id='form-dialog-title'>{t('AddMagnetOrLink')}</DialogTitle>

      <DialogContent>
        <TextField onChange={inputTitle} margin='dense' id='title' label={t('Title')} type='text' fullWidth />
        <TextField onChange={inputPoster} margin='dense' id='poster' label={t('Poster')} type='url' fullWidth />
        <TextField
          onChange={inputMagnet}
          autoFocus
          margin='dense'
          id='magnet'
          label={t('MagnetOrTorrentFileLink')}
          type='text'
          fullWidth
        />
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} color='primary' variant='outlined'>
          {t('Cancel')}
        </Button>

        <Button variant='contained' disabled={!link} onClick={handleSave} color='primary'>
          {t('Add')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
