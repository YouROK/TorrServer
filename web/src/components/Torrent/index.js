import { useEffect, useRef, useState } from 'react'
import Button from '@material-ui/core/Button'

import 'fontsource-roboto'

import HeightIcon from '@material-ui/icons/Height';
import CloseIcon from '@material-ui/icons/Close';
import DeleteIcon from '@material-ui/icons/Delete'
import DialogActions from '@material-ui/core/DialogActions'
import Dialog from '@material-ui/core/Dialog'

import { getPeerString, humanizeSize } from '../../utils/Utils'

import DialogTorrentInfo from '../DialogTorrentInfo'
import { torrentsHost } from '../../utils/Hosts'
import DialogCacheInfo from '../DialogCacheInfo'
import DataUsageIcon from '@material-ui/icons/DataUsage'
import { NoImageIcon } from '../../icons';
import { StyledButton, TorrentCard, TorrentCardButtons, TorrentCardDescription, TorrentCardDescriptionContent, TorrentCardDescriptionLabel, TorrentCardPoster } from './style';

export default function Torrent(props) {
    const [open, setOpen] = useState(false)
    const [showCache, setShowCache] = useState(false)
    const [torrent, setTorrent] = useState(props.torrent)
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

    const { title, name, poster, torrent_size, download_speed } = torrent

    return (
        <>

            <TorrentCard>
                <TorrentCardPoster isPoster={poster}>
                    {poster
                        ? <img src={poster} alt="poster" />
                        : <NoImageIcon />}
                </TorrentCardPoster>

                <TorrentCardButtons>
                    <StyledButton
                        onClick={() => {
                            setShowCache(true)
                            setOpen(true)
                        }}
                    >
                        <DataUsageIcon />
                        Cache
                    </StyledButton>

                    <StyledButton
                        onClick={() => dropTorrent(torrent)}
                    >
                        <CloseIcon />
                        Drop
                    </StyledButton>

                    <StyledButton
                        onClick={() => deleteTorrent(torrent)}
                    >
                        <DeleteIcon />
                        Delete
                    </StyledButton>

                    <StyledButton
                        onClick={() => {
                            setShowCache(false)
                            setOpen(true)
                        }}
                    >
                        <HeightIcon />
                        Details
                    </StyledButton>
                </TorrentCardButtons>

                <TorrentCardDescription>
                    <TorrentCardDescriptionLabel>Name</TorrentCardDescriptionLabel>
                    <TorrentCardDescriptionContent>{title || name}</TorrentCardDescriptionContent>

                    <TorrentCardDescriptionLabel>Size</TorrentCardDescriptionLabel>
                    <TorrentCardDescriptionContent>{torrent_size > 0 && humanizeSize(torrent_size)}</TorrentCardDescriptionContent>

                    <TorrentCardDescriptionLabel>Download speed</TorrentCardDescriptionLabel>
                    <TorrentCardDescriptionContent>{download_speed > 0 ? humanizeSize(download_speed) : '---'}</TorrentCardDescriptionContent>

                    <TorrentCardDescriptionLabel>Peers</TorrentCardDescriptionLabel>
                    <TorrentCardDescriptionContent>{getPeerString(torrent) || '---'}</TorrentCardDescriptionContent>
                </TorrentCardDescription>
            </TorrentCard>

            <Dialog
                open={open}
                onClose={() => setOpen(false)}
                aria-labelledby="form-dialog-title"
                fullWidth
                maxWidth={'lg'}
            >
                {!showCache ? <DialogTorrentInfo torrent={(open, torrent)} /> : <DialogCacheInfo hash={(open, torrent.hash)} />}
                <DialogActions>
                    <Button
                        variant="outlined"
                        color="primary"
                        onClick={() => setOpen(false)}
                    >
                        OK
                    </Button>
                </DialogActions>
            </Dialog>
        </>
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
