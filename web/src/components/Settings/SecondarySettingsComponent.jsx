import { useTranslation } from 'react-i18next'
import TextField from '@material-ui/core/TextField'
import { FormControlLabel, InputLabel, Select, Switch } from '@material-ui/core'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function SecondarySettingsComponent({ settings, inputForm }) {
  const { t } = useTranslation()

  const {
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
  } = settings || {}

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('SettingsDialog.AdditionalSettings')}</SettingSectionLabel>

      <FormControlLabel
        control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='secondary' />}
        label='IPv6'
      />
      <br />
      <FormControlLabel
        control={<Switch checked={!DisableTCP} onChange={inputForm} id='DisableTCP' color='secondary' />}
        label='TCP (Transmission Control Protocol)'
      />
      <br />
      <FormControlLabel
        control={<Switch checked={!DisableUTP} onChange={inputForm} id='DisableUTP' color='secondary' />}
        label='μTP (Micro Transport Protocol)'
      />
      <br />
      <FormControlLabel
        control={<Switch checked={!DisablePEX} onChange={inputForm} id='DisablePEX' color='secondary' />}
        label='PEX (Peer Exchange)'
      />
      <br />
      <FormControlLabel
        control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='secondary' />}
        label={t('SettingsDialog.ForceEncrypt')}
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='TorrentDisconnectTimeout'
        label={t('SettingsDialog.TorrentDisconnectTimeout')}
        value={TorrentDisconnectTimeout}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='ConnectionsLimit'
        label={t('SettingsDialog.ConnectionsLimit')}
        value={ConnectionsLimit}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <FormControlLabel
        control={<Switch checked={!DisableDHT} onChange={inputForm} id='DisableDHT' color='secondary' />}
        label={t('SettingsDialog.DHT')}
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='DhtConnectionLimit'
        label={t('SettingsDialog.DhtConnectionLimit')}
        value={DhtConnectionLimit}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='DownloadRateLimit'
        label={t('SettingsDialog.DownloadRateLimit')}
        value={DownloadRateLimit}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <FormControlLabel
        control={<Switch checked={!DisableUpload} onChange={inputForm} id='DisableUpload' color='secondary' />}
        label={t('SettingsDialog.Upload')}
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='UploadRateLimit'
        label={t('SettingsDialog.UploadRateLimit')}
        value={UploadRateLimit}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='PeersListenPort'
        label={t('SettingsDialog.PeersListenPort')}
        value={PeersListenPort}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <FormControlLabel
        control={<Switch checked={!DisableUPNP} onChange={inputForm} id='DisableUPNP' color='secondary' />}
        label='UPnP (Universal Plug and Play)'
      />
      <br />
      <InputLabel htmlFor='RetrackersMode'>{t('SettingsDialog.RetrackersMode')}</InputLabel>
      <Select
        onChange={inputForm}
        margin='normal'
        type='number'
        native
        id='RetrackersMode'
        value={RetrackersMode}
        variant='outlined'
      >
        <option value={0}>{t('SettingsDialog.DontAddRetrackers')}</option>
        <option value={1}>{t('SettingsDialog.AddRetrackers')}</option>
        <option value={2}>{t('SettingsDialog.RemoveRetrackers')}</option>
        <option value={3}>{t('SettingsDialog.ReplaceRetrackers')}</option>
      </Select>
      <br />
    </SecondarySettingsContent>
  )
}
