import React, { useEffect, useRef } from 'react'
import Container from '@material-ui/core/Container'
import Torrent from './Torrent'
import List from '@material-ui/core/List'
import { Typography } from '@material-ui/core'
import { torrentsHost } from '../utils/Hosts'

export default function TorrentList(props, onChange) {
    const [torrents, setTorrents] = React.useState([])
    const [offline, setOffline] = React.useState(true)
    const timerID = useRef(-1)

    useEffect(() => {
        timerID.current = setInterval(() => {
            getTorrentList((torrs) => {
                if (torrs) setOffline(false)
                else setOffline(true)
                setTorrents(torrs)
            })
        }, 1000)

        return () => {
            clearInterval(timerID.current)
        }
    }, [])

    return (
        <React.Fragment>
            <Container maxWidth="lg">{!offline ? <List>{torrents && torrents.map((torrent) => <Torrent key={torrent.hash} torrent={torrent} />)}</List> : <Typography>Offline</Typography>}</Container>
        </React.Fragment>
    )
}

function getTorrentList(callback) {
    fetch(torrentsHost(), {
        method: 'post',
        body: JSON.stringify({ action: 'list' }),
        headers: {
            Accept: 'application/json, text/plain, */*',
            'Content-Type': 'application/json',
        },
    })
        .then((res) => res.json())
        .then(
            (json) => {
                callback(json)
            },
            (error) => {
                callback(null)
            }
        )
}
