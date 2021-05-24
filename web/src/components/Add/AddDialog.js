import { useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import { torrentsHost } from '../../utils/Hosts'

export default function AddDialog({ handleClose }) {
    const [magnet, setMagnet] = useState('')
    const [title, setTitle] = useState('')
    const [poster, setPoster] = useState('')

    const inputMagnet = ({ target: { value } }) => setMagnet(value)
    const inputTitle = ({ target: { value } }) => setTitle(value)
    const inputPoster = ({ target: { value } }) => setPoster(value)

    const handleCloseSave = () => {
        try {
            if (!magnet) return

            fetch(torrentsHost(), {
                method: 'post',
                body: JSON.stringify({
                    action: 'add',
                    link: magnet,
                    title: title,
                    poster: poster,
                    save_to_db: true,
                }),
                headers: {
                    Accept: 'application/json, text/plain, */*',
                    'Content-Type': 'application/json',
                },
            })
            handleClose()
        } catch (e) {
            console.log(e)
        }
    }

    return (
        <Dialog open onClose={handleClose} aria-labelledby="form-dialog-title" fullWidth>
            <DialogTitle id="form-dialog-title">Add magnet or link to torrent file</DialogTitle>

            <DialogContent>
                <TextField onChange={inputTitle} margin="dense" id="title" label="Title" type="text" fullWidth />
                <TextField onChange={inputPoster} margin="dense" id="poster" label="Poster" type="url" fullWidth />
                <TextField onChange={inputMagnet} autoFocus margin="dense" id="magnet" label="Magnet or torrent file link" type="text" fullWidth />
            </DialogContent>

            <DialogActions>
                <Button onClick={handleClose} color="primary" variant="outlined">
                    Cancel
				</Button>

                <Button disabled={!magnet} onClick={handleCloseSave} color="primary" variant="outlined">
                    Add
				</Button>
            </DialogActions>
        </Dialog>
    )
}