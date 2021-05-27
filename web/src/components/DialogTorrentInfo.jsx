import { useEffect, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import { Button, ButtonGroup, Grid, List, ListItem } from '@material-ui/core'
import CachedIcon from '@material-ui/icons/Cached'
import LinearProgress from '@material-ui/core/LinearProgress'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { playlistTorrHost, streamHost, viewedHost } from 'utils/Hosts'

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
    borderRadius: '5px',
  },
}

export default function DialogTorrentInfo({ torrent, open }) {
  const [torrentLocalComponentValue, setTorrentLocalComponentValue] = useState(torrent)
  const [viewed, setViewed] = useState(null)
  const [progress, setProgress] = useState(-1)

  useEffect(() => {
    setTorrentLocalComponentValue(torrent)
    if (torrentLocalComponentValue.stat === 2)
      setProgress((torrentLocalComponentValue.preloaded_bytes * 100) / torrentLocalComponentValue.preload_size)
    getViewed(torrent.hash, list => {
      if (list) {
        const lst = list.map(itm => itm.file_index)
        setViewed(lst)
      } else setViewed(null)
    })
  }, [torrent, open])

  return (
    <div>
      <DialogTitle id='form-dialog-title'>
        <Grid container spacing={1}>
          <Grid item>
            {torrentLocalComponentValue.poster && (
              <img alt='' height='200' align='left' style={style.poster} src={torrentLocalComponentValue.poster} />
            )}
          </Grid>
          <Grid style={style.width80} item>
            {torrentLocalComponentValue.title}{' '}
            {torrentLocalComponentValue.name &&
              torrentLocalComponentValue.name !== torrentLocalComponentValue.title &&
              ` | ${torrentLocalComponentValue.name}`}
            <Typography>
              <b>Peers: </b> {getPeerString(torrentLocalComponentValue)}
              <br />
              <b>Loaded: </b> {getPreload(torrentLocalComponentValue)}
              <br />
              <b>Speed: </b> {humanizeSize(torrentLocalComponentValue.download_speed)}
              <br />
              <b>Status: </b> {torrentLocalComponentValue.stat_string}
              <br />
            </Typography>
          </Grid>
        </Grid>
        {torrentLocalComponentValue.stat === 2 && (
          <LinearProgress style={{ marginTop: '10px' }} variant='determinate' value={progress} />
        )}
      </DialogTitle>

      <DialogContent>
        <List>
          <ListItem key='TorrentMenu'>
            <ButtonGroup
              style={style.width100}
              variant='contained'
              color='primary'
              aria-label='contained primary button group'
            >
              <Button
                style={style.width100}
                href={`${playlistTorrHost()}/${encodeURIComponent(
                  torrentLocalComponentValue.name || torrentLocalComponentValue.title || 'file',
                )}.m3u?link=${torrentLocalComponentValue.hash}&m3u`}
              >
                Playlist
              </Button>
              <Button
                style={style.width100}
                href={`${playlistTorrHost()}/${encodeURIComponent(
                  torrentLocalComponentValue.name || torrentLocalComponentValue.title || 'file',
                )}.m3u?link=${torrentLocalComponentValue.hash}&m3u&fromlast`}
              >
                Playlist after last view
              </Button>
              <Button
                style={style.width100}
                onClick={() => {
                  remViews(torrentLocalComponentValue.hash)
                  setViewed(null)
                }}
              >
                Remove views
              </Button>
            </ButtonGroup>
          </ListItem>

          {getPlayableFile(torrentLocalComponentValue) &&
            getPlayableFile(torrentLocalComponentValue).map(file => (
              <ButtonGroup style={style.width100} disableElevation variant='contained' color='primary'>
                <Button
                  style={style.width100}
                  href={`${streamHost()}/${encodeURIComponent(file.path.split('\\').pop().split('/').pop())}?link=${
                    torrentLocalComponentValue.hash
                  }&index=${file.id}&play`}
                >
                  <Typography>
                    {file.path.split('\\').pop().split('/').pop()} | {humanizeSize(file.length)}{' '}
                    {viewed && viewed.indexOf(file.id) !== -1 && '| ✓'}
                  </Typography>
                </Button>
                <Button
                  onClick={() =>
                    fetch(`${streamHost()}?link=${torrentLocalComponentValue.hash}&index=${file.id}&preload`)
                  }
                >
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

function remViews(hash) {
  try {
    if (hash)
      fetch(viewedHost(), {
        method: 'post',
        body: JSON.stringify({ action: 'rem', hash, file_index: -1 }),
        headers: {
          Accept: 'application/json, text/plain, */*',
          'Content-Type': 'application/json',
        },
      })
  } catch (e) {
    console.error(e)
  }
}

function getViewed(hash, callback) {
  try {
    fetch(viewedHost(), {
      method: 'post',
      body: JSON.stringify({ action: 'list', hash }),
      headers: {
        Accept: 'application/json, text/plain, */*',
        'Content-Type': 'application/json',
      },
    })
      .then(res => res.json())
      .then(callback)
  } catch (e) {
    console.error(e)
  }
}

function getPlayableFile(torrent) {
  if (!torrent || !torrent.file_stats) return null
  return torrent.file_stats.filter(file => extPlayable.includes(getExt(file.path)))
}

function getExt(filename) {
  const ext = filename.split('.').pop()
  if (ext === filename) return ''
  return ext.toLowerCase()
}

function getPreload(torrent) {
  if (torrent.preloaded_bytes > 0 && torrent.preload_size > 0 && torrent.preloaded_bytes < torrent.preload_size) {
    const progress = ((torrent.preloaded_bytes * 100) / torrent.preload_size).toFixed(2)
    return `${humanizeSize(torrent.preloaded_bytes)} / ${humanizeSize(torrent.preload_size)}   ${progress}%`
  }

  if (!torrent.preloaded_bytes) return humanizeSize(0)

  return humanizeSize(torrent.preloaded_bytes)
}

const extPlayable = [
  // video
  '3g2',
  '3gp',
  'aaf',
  'asf',
  'avchd',
  'avi',
  'drc',
  'flv',
  'iso',
  'm2v',
  'm2ts',
  'm4p',
  'm4v',
  'mkv',
  'mng',
  'mov',
  'mp2',
  'mp4',
  'mpe',
  'mpeg',
  'mpg',
  'mpv',
  'mxf',
  'nsv',
  'ogg',
  'ogv',
  'ts',
  'qt',
  'rm',
  'rmvb',
  'roq',
  'svi',
  'vob',
  'webm',
  'wmv',
  'yuv',
  // audio
  'aac',
  'aiff',
  'ape',
  'au',
  'flac',
  'gsm',
  'it',
  'm3u',
  'm4a',
  'mid',
  'mod',
  'mp3',
  'mpa',
  'pls',
  'ra',
  's3m',
  'sid',
  'wav',
  'wma',
  'xm',
]
