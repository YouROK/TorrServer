import axios from 'axios'
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
import { useTranslation } from 'react-i18next'
import { ThemeProvider } from '@material-ui/core/styles'
import { lightTheme } from 'components/App'

export default function SettingsDialog() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [settings, setSets] = useState({})
  const [show, setShow] = useState(false)
  const [tsHost, setTSHost] = useState(getTorrServerHost())

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

  const onInputHost = ({ target: { value } }) => {
    const host = value.replace(/\/$/gi, '')
    setTorrServerHost(host)
    setTSHost(host)
  }

  const inputForm = ({ target: { type, value, checked, id } }) => {
    const sets = JSON.parse(JSON.stringify(settings))
    if (type === 'number' || type === 'select-one') {
      sets[id] = Number(value)
    } else if (type === 'checkbox') {
      if (
        id === 'DisableTCP' ||
        id === 'DisableUTP' ||
        id === 'DisableUPNP' ||
        id === 'DisableDHT' ||
        id === 'DisablePEX' ||
        id === 'DisableUpload'
      )
        sets[id] = Boolean(!checked)
      else sets[id] = Boolean(checked)
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
      <ListItem button key={t('Settings')} onClick={handleClickOpen}>
        <ListItemIcon>
          <SettingsIcon />
        </ListItemIcon>
        <ListItemText primary={t('Settings')} />
      </ListItem>

      <ThemeProvider theme={lightTheme}>
        <Dialog open={open} onClose={handleClose} aria-labelledby='form-dialog-title' fullWidth>
          <DialogTitle id='form-dialog-title'>{t('Settings')}</DialogTitle>
          <DialogContent>
            <TextField
              onChange={onInputHost}
              margin='dense'
              id='TorrServerHost'
              label={t('Host')}
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
                  label={t('CacheSize')}
                  value={CacheSize}
                  type='number'
                  fullWidth
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='ReaderReadAHead'
                  label={t('ReaderReadAHead')}
                  value={ReaderReadAHead}
                  type='number'
                  fullWidth
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={PreloadBuffer} onChange={inputForm} id='PreloadBuffer' color='primary' />}
                  label={t('PreloadBuffer')}
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={UseDisk} onChange={inputForm} id='UseDisk' color='primary' />}
                  label={t('UseDisk')}
                />
                <br />
                <small>{t('UseDiskDesc')}</small>
                <br />
                <FormControlLabel
                  control={
                    <Switch checked={RemoveCacheOnDrop} onChange={inputForm} id='RemoveCacheOnDrop' color='primary' />
                  }
                  label={t('RemoveCacheOnDrop')}
                />
                <br />
                <small>{t('RemoveCacheOnDropDesc')}</small>
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='TorrentsSavePath'
                  label={t('TorrentsSavePath')}
                  value={TorrentsSavePath}
                  type='url'
                  fullWidth
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='primary' />}
                  label={t('EnableIPv6')}
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={!DisableTCP} onChange={inputForm} id='DisableTCP' color='primary' />}
                  label={t('TCP')}
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={!DisableUTP} onChange={inputForm} id='DisableUTP' color='primary' />}
                  label={t('UTP')}
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={!DisablePEX} onChange={inputForm} id='DisablePEX' color='primary' />}
                  label={t('PEX')}
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='primary' />}
                  label={t('ForceEncrypt')}
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='TorrentDisconnectTimeout'
                  label={t('TorrentDisconnectTimeout')}
                  value={TorrentDisconnectTimeout}
                  type='number'
                  fullWidth
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='ConnectionsLimit'
                  label={t('ConnectionsLimit')}
                  value={ConnectionsLimit}
                  type='number'
                  fullWidth
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={!DisableDHT} onChange={inputForm} id='DisableDHT' color='primary' />}
                  label={t('DHT')}
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='DhtConnectionLimit'
                  label={t('DhtConnectionLimit')}
                  value={DhtConnectionLimit}
                  type='number'
                  fullWidth
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='DownloadRateLimit'
                  label={t('DownloadRateLimit')}
                  value={DownloadRateLimit}
                  type='number'
                  fullWidth
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={!DisableUpload} onChange={inputForm} id='DisableUpload' color='primary' />}
                  label={t('Upload')}
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='UploadRateLimit'
                  label={t('UploadRateLimit')}
                  value={UploadRateLimit}
                  type='number'
                  fullWidth
                />
                <br />
                <TextField
                  onChange={inputForm}
                  margin='dense'
                  id='PeersListenPort'
                  label={t('PeersListenPort')}
                  value={PeersListenPort}
                  type='number'
                  fullWidth
                />
                <br />
                <FormControlLabel
                  control={<Switch checked={!DisableUPNP} onChange={inputForm} id='DisableUPNP' color='primary' />}
                  label={t('UPNP')}
                />
                <br />
                <InputLabel htmlFor='RetrackersMode'>{t('RetrackersMode')}</InputLabel>
                <Select onChange={inputForm} type='number' native id='RetrackersMode' value={RetrackersMode}>
                  <option value={0}>{t('DontAddRetrackers')}</option>
                  <option value={1}>{t('AddRetrackers')}</option>
                  <option value={2}>{t('RemoveRetrackers')}</option>
                  <option value={3}>{t('ReplaceRetrackers')}</option>
                </Select>
                <br />
              </>
            )}
          </DialogContent>

          <DialogActions>
            <Button onClick={handleClose} color='primary' variant='outlined'>
              {t('Cancel')}
            </Button>

            <Button onClick={handleSave} color='primary' variant='outlined'>
              {t('Save')}
            </Button>
          </DialogActions>
        </Dialog>
      </ThemeProvider>
    </div>
  )
}
