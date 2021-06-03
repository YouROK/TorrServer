import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useEffect, useState } from 'react'
import SettingsIcon from '@material-ui/icons/Settings'
import Dialog from '@material-ui/core/Dialog'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import TextField from '@material-ui/core/TextField'
import DialogActions from '@material-ui/core/DialogActions'
import Button from '@material-ui/core/Button'
import { FormControlLabel, InputLabel, Select, Switch } from '@material-ui/core'
import { settingsHost, setTorrServerHost, getTorrServerHost } from 'utils/Hosts'
import axios from 'axios'

export default function SettingsDialog() {
  const [open, setOpen] = useState(false)
  const [settings, setSets] = useState({})
  const [show, setShow] = useState(false)

  const { protocol, hostname, port } = window.location
  const [tsHost, setTSHost] = useState(getTorrServerHost() || `${protocol}//${hostname}${port ? `:${port}` : ''}`)

  const handleClickOpen = () => setOpen(true)
  const handleClose = () => setOpen(false)
  const handleSave = () => {
    setOpen(false)
    const sets = JSON.parse(JSON.stringify(settings))
    sets.CacheSize *= 1024 * 1024
    axios.post(settingsHost(), { action: 'set', sets })
  }

  useEffect(() => {
    axios
      .post(settingsHost(), { action: 'get' })
      .then(({ data }) => {
        setSets({ ...data, CacheSize: data.CacheSize / (1024 * 1024) })
        setShow(true)
      })
      .catch(() => setShow(false))
  }, [tsHost])

  const onInputHost = event => {
    const host = event.target.value
    setTorrServerHost(host)
    setTSHost(host)
  }

  const inputForm = ({ target: { type, value, checked, id } }) => {
    const sets = JSON.parse(JSON.stringify(settings))
    if (type === 'number' || type === 'select-one') {
      sets[id] = Number(value)
    } else if (type === 'checkbox') {
      sets[id] = Boolean(checked)
    } else if (type === 'url') {
      sets[id] = value
    }
    setSets(sets)
  }

  const {
    CacheSize,
    PreloadBuffer,
    ReaderReadAHead,
    RetrackersMode,
    TorrentDisconnectTimeout,
    EnableIPv6,
    ForceEncrypt,
    DisableTCP,
    DisableUTP,
    DisableUPNP,
    DisableDHT,
    DisablePEX,
    DisableUpload,
    DownloadRateLimit,
    UploadRateLimit,
    ConnectionsLimit,
    DhtConnectionLimit,
    PeersListenPort,
    UseDisk,
    TorrentsSavePath,
    RemoveCacheOnDrop,
  } = settings

  return (
    <div>
      <ListItem button key='Settings' onClick={handleClickOpen}>
        <ListItemIcon>
          <SettingsIcon />
        </ListItemIcon>
        <ListItemText primary='Settings' />
      </ListItem>

      <Dialog open={open} onClose={handleClose} aria-labelledby='form-dialog-title' fullWidth>
        <DialogTitle id='form-dialog-title'>Settings</DialogTitle>
        <DialogContent>
          <TextField
            onChange={onInputHost}
            margin='dense'
            id='TorrServerHost'
            label='Host'
            value={tsHost}
            type='url'
            fullWidth
          />
          {show && (
            <>
              <TextField
                onChange={inputForm}
                margin='dense'
                id='CacheSize'
                label='Cache size'
                value={CacheSize}
                type='number'
                fullWidth
              />
              <FormControlLabel
                control={<Switch checked={PreloadBuffer} onChange={inputForm} id='PreloadBuffer' color='primary' />}
                label='Preload buffer'
              />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='ReaderReadAHead'
                label='Reader readahead'
                value={ReaderReadAHead}
                type='number'
                fullWidth
              />
              <br />
              <br />
              <InputLabel htmlFor='RetrackersMode'>Retracker mode</InputLabel>
              <Select onChange={inputForm} type='number' native id='RetrackersMode' value={RetrackersMode}>
                <option value={0}>Don&apos;t add retrackers</option>
                <option value={1}>Add retrackers</option>
                <option value={2}>Remove retrackers</option>
                <option value={3}>Replace retrackers</option>
              </Select>
              <TextField
                onChange={inputForm}
                margin='dense'
                id='TorrentDisconnectTimeout'
                label='Torrent disconnect timeout'
                value={TorrentDisconnectTimeout}
                type='number'
                fullWidth
              />
              <FormControlLabel
                control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='primary' />}
                label='Enable IPv6'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='primary' />}
                label='Force encrypt'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={DisableTCP} onChange={inputForm} id='DisableTCP' color='primary' />}
                label='Disable TCP'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={DisableUTP} onChange={inputForm} id='DisableUTP' color='primary' />}
                label='Disable UTP'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={DisableUPNP} onChange={inputForm} id='DisableUPNP' color='primary' />}
                label='Disable UPNP'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={DisableDHT} onChange={inputForm} id='DisableDHT' color='primary' />}
                label='Disable DHT'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={DisablePEX} onChange={inputForm} id='DisablePEX' color='primary' />}
                label='Disable PEX'
              />
              <br />
              <FormControlLabel
                control={<Switch checked={DisableUpload} onChange={inputForm} id='DisableUpload' color='primary' />}
                label='Disable upload'
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='DownloadRateLimit'
                label='Download rate limit'
                value={DownloadRateLimit}
                type='number'
                fullWidth
              />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='UploadRateLimit'
                label='Upload rate limit'
                value={UploadRateLimit}
                type='number'
                fullWidth
              />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='ConnectionsLimit'
                label='Connections limit'
                value={ConnectionsLimit}
                type='number'
                fullWidth
              />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='DhtConnectionLimit'
                label='Dht connection limit'
                value={DhtConnectionLimit}
                type='number'
                fullWidth
              />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='PeersListenPort'
                label='Peers listen port'
                value={PeersListenPort}
                type='number'
                fullWidth
              />
              <br />
              <FormControlLabel
                control={<Switch checked={UseDisk} onChange={inputForm} id='UseDisk' color='primary' />}
                label='Use disk'
              />
              <br />
              <FormControlLabel
                  control={<Switch checked={RemoveCacheOnDrop} onChange={inputForm} id='RemoveCacheOnDrop' color='primary' />}
                  label='Remove cache from disk on drop torrent'
              />
              <br/>
              <small>If disabled, remove cache on delete torrent</small>
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='TorrentsSavePath'
                label='Torrents save path'
                value={TorrentsSavePath}
                type='url'
                fullWidth
              />
            </>
          )}
        </DialogContent>

        <DialogActions>
          <Button onClick={handleClose} color='primary' variant='outlined'>
            Cancel
          </Button>

          <Button onClick={handleSave} color='primary' variant='outlined'>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}
