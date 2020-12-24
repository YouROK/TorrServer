import React from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogTitle from '@material-ui/core/DialogTitle'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import LibraryAddIcon from '@material-ui/icons/LibraryAdd'
import ListItemText from '@material-ui/core/ListItemText'
import ListItem from '@material-ui/core/ListItem'
import { torrentsHost } from '../utils/Hosts'

export default function AddDialog() {
    const [open, setOpen] = React.useState(false)

    const [magnet, setMagnet] = React.useState('')
    const [title, setTitle] = React.useState('')
    const [poster, setPoster] = React.useState('')

    const handleClickOpen = () => {
        setOpen(true)
    }

    const inputMagnet = (event) => {
        setMagnet(event.target.value)
    }

    const inputTitle = (event) => {
        setTitle(event.target.value)
    }

    const inputPoster = (event) => {
        setPoster(event.target.value)
    }

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
            setOpen(false)
        } catch (e) {
            console.log(e)
        }
    }

    const handleClose = () => {
        setOpen(false)
    }

    return (
        <div>
            <ListItem button key="Add" onClick={handleClickOpen}>
                <ListItemIcon>
                    <LibraryAddIcon />
                </ListItemIcon>
                <ListItemText primary="Add" />
            </ListItem>
            <Dialog open={open} onClose={handleClose} aria-labelledby="form-dialog-title" fullWidth={true}>
                <DialogTitle id="form-dialog-title">Add Magnet</DialogTitle>
                <DialogContent>
                    <DialogContentText>Add magnet or link to torrent file:</DialogContentText>
                    <TextField onChange={inputTitle} margin="dense" id="title" label="Title" type="text" fullWidth />
                    <TextField onChange={inputPoster} margin="dense" id="poster" label="Poster" type="url" fullWidth />
                    <TextField onChange={inputMagnet} autoFocus margin="dense" id="magnet" label="Magnet" type="text" fullWidth />
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose} color="primary" variant="outlined">
                        Cancel
                    </Button>
                    <Button onClick={handleCloseSave} color="primary" variant="outlined">
                        Add
                    </Button>
                </DialogActions>
            </Dialog>
        </div>
    )
}
