import React from 'react'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import DeleteIcon from '@material-ui/icons/Delete'
import { torrentsHost } from '../utils/Hosts'

const fnRemoveAll = () => {
    fetch(torrentsHost(), {
        method: 'post',
        body: JSON.stringify({ action: 'list' }),
        headers: {
            Accept: 'application/json, text/plain, */*',
            'Content-Type': 'application/json',
        },
    })
        .then((res) => res.json())
        .then((json) => {
            json.forEach((torr) => {
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
    return (
        <ListItem button key="Remove all" onClick={fnRemoveAll}>
            <ListItemIcon>
                <DeleteIcon />
            </ListItemIcon>
            <ListItemText primary="Remove all" />
        </ListItem>
    )
}
