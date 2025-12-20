import { useTranslation } from 'react-i18next'
import TextField from '@material-ui/core/TextField'
import {
  FormControlLabel,
  FormGroup,
  FormHelperText,
  InputAdornment,
  InputLabel,
  Select,
  Switch,
} from '@material-ui/core'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function SecondarySettingsComponent({ settings, inputForm }) {
  const { t } = useTranslation()

  const {
    RetrackersMode,
    TorrentDisconnectTimeout,
    EnableDebug,
    EnableDLNA,
    EnableIPv6,
    FriendlyName,
    EnableRutorSearch,
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
    PeersListenPort,
    ResponsiveMode,
    SslPort,
    SslCert,
    SslKey,
    ShowFSActiveTorr,
    // FUSEPath,
  } = settings || {}

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('SettingsDialog.AdditionalSettings')}</SettingSectionLabel>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='secondary' />}
          label='IPv6'
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.EnableIPv6Hint')}</FormHelperText>
      </FormGroup>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!DisableTCP} onChange={inputForm} id='DisableTCP' color='secondary' />}
          label='TCP (Transmission Control Protocol)'
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.DisableTCPHint')}</FormHelperText>
      </FormGroup>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!DisableUTP} onChange={inputForm} id='DisableUTP' color='secondary' />}
          label='μTP (Micro Transport Protocol)'
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.DisableUTPHint')}</FormHelperText>
      </FormGroup>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!DisablePEX} onChange={inputForm} id='DisablePEX' color='secondary' />}
          label='PEX (Peer Exchange)'
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.DisablePEXHint')}</FormHelperText>
      </FormGroup>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='secondary' />}
          label={t('SettingsDialog.ForceEncrypt')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.ForceEncryptHint')}</FormHelperText>
      </FormGroup>
      <TextField
        onChange={inputForm}
        margin='normal'
        id='TorrentDisconnectTimeout'
        label={t('SettingsDialog.TorrentDisconnectTimeout')}
        InputProps={{
          endAdornment: <InputAdornment position='end'>{t('Seconds')}</InputAdornment>,
        }}
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
        helperText={t('SettingsDialog.ConnectionsLimitHint')}
        value={ConnectionsLimit}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!DisableDHT} onChange={inputForm} id='DisableDHT' color='secondary' />}
          label={t('SettingsDialog.DHT')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.DisableDHTHint')}</FormHelperText>
      </FormGroup>
      <TextField
        onChange={inputForm}
        margin='normal'
        id='DownloadRateLimit'
        label={t('SettingsDialog.DownloadRateLimit')}
        InputProps={{
          endAdornment: <InputAdornment position='end'>{t('Kilobytes')}</InputAdornment>,
        }}
        value={DownloadRateLimit}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!DisableUpload} onChange={inputForm} id='DisableUpload' color='secondary' />}
          label={t('SettingsDialog.Upload')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.UploadHint')}</FormHelperText>
      </FormGroup>
      <TextField
        onChange={inputForm}
        margin='normal'
        id='UploadRateLimit'
        label={t('SettingsDialog.UploadRateLimit')}
        InputProps={{
          endAdornment: <InputAdornment position='end'>{t('Kilobytes')}</InputAdornment>,
        }}
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
        helperText={t('SettingsDialog.PeersListenPortHint')}
        value={PeersListenPort}
        type='number'
        variant='outlined'
        fullWidth
      />
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!DisableUPNP} onChange={inputForm} id='DisableUPNP' color='secondary' />}
          label='UPnP (Universal Plug and Play)'
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.DisableUPNPHint')}</FormHelperText>
      </FormGroup>
      <FormControlLabel
        control={<Switch checked={EnableDLNA} onChange={inputForm} id='EnableDLNA' color='secondary' />}
        label={t('SettingsDialog.DLNA')}
        labelPlacement='start'
      />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='FriendlyName'
        label={t('SettingsDialog.FriendlyName')}
        helperText={t('SettingsDialog.FriendlyNameHint')}
        value={FriendlyName}
        type='text'
        variant='outlined'
        fullWidth
      />
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={EnableRutorSearch} onChange={inputForm} id='EnableRutorSearch' color='secondary' />}
          label={t('SettingsDialog.EnableRutorSearch')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.EnableRutorSearchHint')}</FormHelperText>
      </FormGroup>
      <FormControlLabel
        control={<Switch checked={EnableDebug} onChange={inputForm} id='EnableDebug' color='secondary' />}
        label={t('SettingsDialog.EnableDebug')}
        labelPlacement='start'
      />
      <FormControlLabel
        control={<Switch checked={ResponsiveMode} onChange={inputForm} id='ResponsiveMode' color='secondary' />}
        label={t('SettingsDialog.ResponsiveMode')}
        labelPlacement='start'
      />
      <br />
      <InputLabel htmlFor='RetrackersMode'>{t('SettingsDialog.RetrackersMode')}</InputLabel>
      <Select
        onChange={inputForm}
        margin='dense'
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
      <TextField
        onChange={inputForm}
        margin='normal'
        id='SslPort'
        label={t('SettingsDialog.SslPort')}
        helperText={t('SettingsDialog.SslPortHint')}
        value={SslPort}
        type='number'
        variant='outlined'
        fullWidth
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='SslCert'
        label={t('SettingsDialog.SslCert')}
        helperText={t('SettingsDialog.SslCertHint')}
        value={SslCert}
        type='url'
        variant='outlined'
        fullWidth
      />
      <br />
      <TextField
        onChange={inputForm}
        margin='normal'
        id='SslKey'
        label={t('SettingsDialog.SslKey')}
        helperText={t('SettingsDialog.SslKeyHint')}
        value={SslKey}
        type='url'
        variant='outlined'
        fullWidth
      />
      <br />
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={ShowFSActiveTorr} onChange={inputForm} id='ShowFSActiveTorr' color='secondary' />}
          label={t('SettingsDialog.ShowFSActiveTorr')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.ShowFSActiveTorrHint')}</FormHelperText>
      </FormGroup>
    </SecondarySettingsContent>
  )
}
