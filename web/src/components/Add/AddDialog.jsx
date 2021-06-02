import { useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import { torrentsHost } from 'utils/Hosts'
import axios from 'axios'

export default function AddDialog({ handleClose }) {
  const [link, setLink] = useState('')
  const [title, setTitle] = useState('')
  const [poster, setPoster] = useState('')

  const inputMagnet = ({ target: { value } }) => setLink(value)
  const inputTitle = ({ target: { value } }) => setTitle(value)
  const inputPoster = ({ target: { value } }) => setPoster(value)

  const handleSave = () => {
    axios.post(torrentsHost(), { action: 'add', link, title, poster, save_to_db: true }).finally(() => handleClose())
  }

  return (
    <Dialog open onClose={handleClose} aria-labelledby='form-dialog-title' fullWidth>
      <DialogTitle id='form-dialog-title'>Add magnet or link to torrent file</DialogTitle>

      <DialogContent>
        <TextField onChange={inputTitle} margin='dense' id='title' label='Title' type='text' fullWidth />
        <TextField onChange={inputPoster} margin='dense' id='poster' label='Poster' type='url' fullWidth />
        <TextField
          onChange={inputMagnet}
          autoFocus
          margin='dense'
          id='magnet'
          label='Magnet or torrent file link'
          type='text'
          fullWidth
        />
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} color='primary' variant='outlined'>
          Cancel
        </Button>

        <Button variant='contained' disabled={!link} onClick={handleSave} color='primary'>
          Add
        </Button>
      </DialogActions>
    </Dialog>
  )
}
