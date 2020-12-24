import React from 'react'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import ListItem from '@material-ui/core/ListItem'
import PublishIcon from '@material-ui/icons/Publish'
import { torrentUploadHost } from '../utils/Hosts'

const classes = {
    input: {
        display: 'none',
    },
}

export default function UploadDialog() {
    const handleCapture = ({ target }) => {
        let data = new FormData()
        data.append('save', 'true')
        for (let i = 0; i < target.files.length; i++) {
            data.append('file' + i, target.files[i])
        }
        fetch(torrentUploadHost(), {
            method: 'POST',
            body: data,
        })
    }

    return (
        <div>
            <input onChange={handleCapture} accept="*/*" type="file" className={classes.input} style={{ display: 'none' }} id="raised-button-file" multiple />
            <label htmlFor="raised-button-file">
                <ListItem button variant="raised" type="submit" component="span" className={classes.button} key="Upload file">
                    <ListItemIcon>
                        <PublishIcon />
                    </ListItemIcon>
                    <ListItemText primary="Upload file" />
                </ListItem>
            </label>
        </div>
    )
}
