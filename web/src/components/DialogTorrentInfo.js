import React, { useEffect } from 'react'
import Typography from '@material-ui/core/Typography'
import { Button, ButtonGroup, Grid, List, ListItem } from '@material-ui/core'
import CachedIcon from '@material-ui/icons/Cached'
import LinearProgress from '@material-ui/core/LinearProgress';

import { getPeerString, humanizeSize } from '../utils/Utils'
import { playlistTorrHost, streamHost, viewedHost } from '../utils/Hosts'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'

const style = {
    width100: {
        width: '100%',
    },
    width80: {
        width: '80%',
    },
    poster: {
        display: 'flex',
        flexDirection: 'row',
        borderRadius:'5px',
    },
}

export default function DialogTorrentInfo(props) {
    const [torrent, setTorrent] = React.useState(props.torrent)
    const [viewed, setViewed] = React.useState(null)
    const [progress, setProgress] = React.useState(-1)

    useEffect(() => {
        setTorrent(props.torrent)
        if(torrent.stat==2)
            setProgress(torrent.preloaded_bytes * 100 / torrent.preload_size)
        getViewed(props.torrent.hash,(list) => {
            if (list) {
                let lst = list.map((itm) => itm.file_index)
                setViewed(lst)
            }else
                setViewed(null)
        })
    }, [props.torrent, props.open])

    return (
        <div>
            <DialogTitle id="form-dialog-title">
                <Grid container spacing={1}>
                    <Grid item>{torrent.poster && <img alt="" height="200" align="left" style={style.poster} src={torrent.poster} />}</Grid>
                    <Grid style={style.width80} item>
                        {torrent.title} {torrent.name && torrent.name !== torrent.title && ' | ' + torrent.name}
                        <Typography>
                            <b>Peers: </b> {getPeerString(torrent)}
                            <br />
                            <b>Loaded: </b> {getPreload(torrent)}
                            <br />
                            <b>Speed: </b> {humanizeSize(torrent.download_speed)}
                            <br />
                            <b>Status: </b> {torrent.stat_string}
                            <br />
                        </Typography>
                    </Grid>
                </Grid>
                {torrent.stat==2 && <LinearProgress style={{marginTop:'10px'}} variant="determinate" value={progress} />}
            </DialogTitle>
            <DialogContent>
                <List>
                    <ListItem>
                        <ButtonGroup style={style.width100} variant="contained" color="primary" aria-label="contained primary button group">
                            <Button style={style.width100} href={playlistTorrHost() + '/' + encodeURIComponent(torrent.name || torrent.title || 'file') + '.m3u?link=' + torrent.hash + '&m3u'}>
                                Playlist
                            </Button>
                            <Button style={style.width100} href={playlistTorrHost() + '/' + encodeURIComponent(torrent.name || torrent.title || 'file') + '.m3u?link=' + torrent.hash + '&m3u&fromlast'}>
                                Playlist after last view
                            </Button>
                        </ButtonGroup>
                    </ListItem>
                    {getPlayableFile(torrent) &&
                        getPlayableFile(torrent).map((file) => (
                            <ButtonGroup style={style.width100} disableElevation variant="contained" color="primary">
                                <Button
                                    style={style.width100}
                                    href={streamHost() + '/' + encodeURIComponent(file.path.split('\\').pop().split('/').pop()) + '?link=' + torrent.hash + '&index=' + file.id + '&play'}
                                >
                                    <Typography>
                                        {file.path.split('\\').pop().split('/').pop()} | {humanizeSize(file.length)} {viewed && viewed.indexOf(file.id)!=-1 && "| ✓"}
                                    </Typography>
                                </Button>
                                <Button onClick={() => fetch(streamHost() + '?link=' + torrent.hash + '&index=' + file.id + '&preload')}>
                                    <CachedIcon />
                                    <Typography>Preload</Typography>
                                </Button>
                            </ButtonGroup>
                        ))}
                </List>
            </DialogContent>
        </div>
    )
}

function getViewed(hash, callback) {
    try {
        fetch(viewedHost(), {
            method: 'post',
            body: JSON.stringify({ action: 'list', hash: hash }),
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
    } catch (e) {
        console.error(e)
    }
}

function getPlayableFile(torrent){
    if (!torrent || !torrent.file_stats)
        return null
    return torrent.file_stats.filter(file => extPlayable.includes(getExt(file.path)))
}

function getExt(filename){
    const ext = filename.split('.').pop()
    if (ext == filename)
        return ''
    return ext.toLowerCase()
}

function getPreload(torrent) {
    if (torrent.preloaded_bytes > 0 && torrent.preload_size > 0 && torrent.preloaded_bytes < torrent.preload_size) {
        let progress = ((torrent.preloaded_bytes * 100) / torrent.preload_size).toFixed(2)
        return humanizeSize(torrent.preloaded_bytes) + ' / ' + humanizeSize(torrent.preload_size) + '   ' + progress + '%'
    }

    if (!torrent.preloaded_bytes) return humanizeSize(0)

    return humanizeSize(torrent.preloaded_bytes)
}

const extPlayable = [
// video
    "3g2",
    "3gp",
    "aaf",
    "asf",
    "avchd",
    "avi",
    "drc",
    "flv",
    "iso",
    "m2v",
    "m2ts",
    "m4p",
    "m4v",
    "mkv",
    "mng",
    "mov",
    "mp2",
    "mp4",
    "mpe",
    "mpeg",
    "mpg",
    "mpv",
    "mxf",
    "nsv",
    "ogg",
    "ogv",
    "ts",
    "qt",
    "rm",
    "rmvb",
    "roq",
    "svi",
    "vob",
    "webm",
    "wmv",
    "yuv",
// audio
    "aac",
    "aiff",
    "ape",
    "au",
    "flac",
    "gsm",
    "it",
    "m3u",
    "m4a",
    "mid",
    "mod",
    "mp3",
    "mpa",
    "pls",
    "ra",
    "s3m",
    "sid",
    "wav",
    "wma",
    "xm"
]
