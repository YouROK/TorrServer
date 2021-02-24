import React, { useEffect, useRef } from 'react'
import ButtonGroup from '@material-ui/core/ButtonGroup'
import Button from '@material-ui/core/Button'

import 'fontsource-roboto'

import DeleteIcon from '@material-ui/icons/Delete'
import Typography from '@material-ui/core/Typography'
import ListItem from '@material-ui/core/ListItem'
import DialogActions from '@material-ui/core/DialogActions'
import Dialog from '@material-ui/core/Dialog'

import { getPeerString, humanizeSize } from '../utils/Utils'

import DialogTorrentInfo from './DialogTorrentInfo'
import { torrentsHost } from '../utils/Hosts'
import DialogCacheInfo from './DialogCacheInfo'
import DataUsageIcon from '@material-ui/icons/DataUsage'

export default function Torrent(props) {
    const [open, setOpen] = React.useState(false)
    const [showCache, setShowCache] = React.useState(false)
    const [torrent, setTorrent] = React.useState(props.torrent)
    const timerID = useRef(-1)

    useEffect(() => {
        setTorrent(props.torrent)
    }, [props.torrent])

    useEffect(() => {
        if (open)
            timerID.current = setInterval(() => {
                getTorrent(torrent.hash, (torr, error) => {
                    if (error) console.error(error)
                    else if (torr) setTorrent(torr)
                })
            }, 1000)
        else clearInterval(timerID.current)

        return () => {
            clearInterval(timerID.current)
        }
    }, [torrent.hash, open])

    return (
        <div>
            <ListItem>
                <ButtonGroup style={{width:'100%',boxShadow:'2px 2px 2px gray'}} disableElevation variant="contained" color="primary">
                    <Button
                        style={{width: '100%', justifyContent:'start'}}
                        onClick={() => {
                            setShowCache(false)
                            setOpen(true)
                        }}
                    >
                        {torrent.poster &&
                            <img src={torrent.poster} alt="" align="left" style={{width: 'auto',height:'100px',margin:'0 10px 0 0',borderRadius:'5px'}}/>
                        }
                        <Typography>
                            {torrent.title ? torrent.title : torrent.name}
                            {torrent.torrent_size > 0 ? ' | ' + humanizeSize(torrent.torrent_size) : ''}
                            {torrent.download_speed > 0 ? ' | ' + humanizeSize(torrent.download_speed) + '/sec' : ''}
                            {getPeerString(torrent) ? ' | ' + getPeerString(torrent) : '' }
                        </Typography>
                    </Button>
                    <Button
                        onClick={() => {
                            setShowCache(true)
                            setOpen(true)
                        }}
                    >
                        <DataUsageIcon />
                        <Typography>Cache</Typography>
                    </Button>
                    <Button
                        onClick={() => {
                            deleteTorrent(torrent)
                        }}
                    >
                        <DeleteIcon />
                        <Typography>Delete</Typography>
                    </Button>
                </ButtonGroup>
            </ListItem>
            <Dialog
                open={open}
                onClose={() => {
                    setOpen(false)
                }}
                aria-labelledby="form-dialog-title"
                fullWidth={true}
                maxWidth={'lg'}
            >
                {!showCache ? <DialogTorrentInfo torrent={(open, torrent)} /> : <DialogCacheInfo hash={(open, torrent.hash)} />}
                <DialogActions>
                    <Button
                        variant="outlined"
                        color="primary"
                        onClick={() => {
                            setOpen(false)
                        }}
                    >
                        OK
                    </Button>
                    <Button
                        variant="outlined"
                        color="primary"
                        onClick={() => {
                            setOpen(false)
                            dropTorrent(torrent)
                        }}
                    >
                        Drop
                    </Button>
                </DialogActions>
            </Dialog>
        </div>
    )
}

function getTorrent(hash, callback) {
    try {
        fetch(torrentsHost(), {
            method: 'post',
            body: JSON.stringify({ action: 'get', hash: hash }),
            headers: {
                Accept: 'application/json, text/plain, */*',
                'Content-Type': 'application/json',
            },
        })
            .then((res) => res.json())
            .then(
                (json) => {
                    callback(json, null)
                },
                (error) => {
                    callback(null, error)
                }
            )
    } catch (e) {
        console.error(e)
    }
}

function deleteTorrent(torrent) {
    try {
        fetch(torrentsHost(), {
            method: 'post',
            body: JSON.stringify({
                action: 'rem',
                hash: torrent.hash,
            }),
            headers: {
                Accept: 'application/json, text/plain, */*',
                'Content-Type': 'application/json',
            },
        })
    } catch (e) {
        console.error(e)
    }
}

function dropTorrent(torrent) {
    try {
        fetch(torrentsHost(), {
            method: 'post',
            body: JSON.stringify({
                action: 'drop',
                hash: torrent.hash,
            }),
            headers: {
                Accept: 'application/json, text/plain, */*',
                'Content-Type': 'application/json',
            },
        })
    } catch (e) {
        console.error(e)
    }
}

/*
{
	"title": "Mulan 2020",
	"poster": "https://kinohod.ru/o/88/d3/88d3054f-8fd3-4daf-8977-bb4bc8b95206.jpg",
	"timestamp": 1606897747,
	"name": "Mulan.2020.MVO.BDRip.1.46Gb",
	"hash": "f6c992b437c04d0f5a44b42852bb61de7ce90f9a",
	"stat": 2,
	"stat_string": "Torrent preload",
	"loaded_size": 6160384,
	"torrent_size": 1569489783,
	"preloaded_bytes": 5046272,
	"preload_size": 20971520,
	"download_speed": 737156.3390754947,
	"total_peers": 149,
	"pending_peers": 136,
	"active_peers": 10,
	"connected_seeders": 9,
	"half_open_peers": 15,
	"bytes_written": 100327,
	"bytes_read": 8077590,
	"bytes_read_data": 7831552,
	"bytes_read_useful_data": 6160384,
	"chunks_read": 478,
	"chunks_read_useful": 376,
	"chunks_read_wasted": 102,
	"pieces_dirtied_good": 2,
	"file_stats": [{
		"id": 1,
		"path": "Mulan.2020.MVO.BDRip.1.46Gb/Mulan.2020.MVO.BDRip.1.46Gb.avi",
		"length": 1569415168
	}, {
		"id": 2,
		"path": "Mulan.2020.MVO.BDRip.1.46Gb/Mulan.2020.MVO.BDRip.1.46Gb_forced.rus.srt",
		"length": 765
	}, {
		"id": 3,
		"path": "Mulan.2020.MVO.BDRip.1.46Gb/Mulan.2020.MVO.BDRip.1.46Gb_full.rus.srt",
		"length": 73850
	}]
}
 */
