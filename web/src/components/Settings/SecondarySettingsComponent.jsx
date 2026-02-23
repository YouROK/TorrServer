import { useTranslation } from 'react-i18next'
import TextField from '@material-ui/core/TextField'
import {
  Box,
  Button,
  CircularProgress,
  FormControlLabel,
  FormGroup,
  FormHelperText,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  Switch,
} from '@material-ui/core'
import { styled } from '@material-ui/core/styles'
import { useEffect, useMemo, useState } from 'react'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

// Create a styled status message component
const StatusMessage = styled('div')(({ theme, severity }) => ({
  padding: theme.spacing(1.5, 2),
  marginTop: theme.spacing(1),
  borderRadius: theme.shape.borderRadius,
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  backgroundColor:
    severity === 'error' ? '#f44336' : severity === 'success' ? '#4caf50' : severity === 'info' ? '#2196f3' : '#ff9800',
  color: 'white',
  '& button': {
    color: 'white',
    minWidth: 'auto',
    padding: '4px 8px',
    marginLeft: theme.spacing(1),
  },
}))

export default function SecondarySettingsComponent({ settings, inputForm }) {
  const { t } = useTranslation()
  const [storageSettings, setStorageSettings] = useState({
    settings: 'json',
    viewed: 'bbolt',
  })
  const [storageStatus, setStorageStatus] = useState({ message: '', type: '' })
  const [loading, setLoading] = useState(false)
  const {
    RetrackersMode,
    TorrentDisconnectTimeout,
    EnableDebug,
    EnableDLNA,
    EnableIPv6,
    FriendlyName,
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
    EnableProxy,
    ProxyHosts,
  } = settings || {}

  // Local state for ProxyHosts text input
  const [proxyHostsText, setProxyHostsText] = useState('')

  // Sync proxyHostsText with ProxyHosts when settings change
  useEffect(() => {
    const textValue = Array.isArray(ProxyHosts) ? ProxyHosts.join(', ') : ProxyHosts || ''
    setProxyHostsText(textValue)
  }, [ProxyHosts])

  // Use useMemo to compute basePath once
  const basePath = useMemo(() => {
    if (typeof window !== 'undefined') {
      return window.location.pathname.split('/')[1] || ''
    }
    return ''
  }, [])

  // Helper function to build API URL
  const getApiUrl = useMemo(
    () => endpoint => {
      const prefix = basePath ? `/${basePath}` : ''
      return `${prefix}${endpoint}`
    },
    [basePath],
  )

  useEffect(() => {
    const loadStorageSettings = async () => {
      try {
        const response = await fetch(getApiUrl('/storage/settings')) // /api/storage/settings
        if (response.ok) {
          const prefs = await response.json()
          setStorageSettings(prefs)
        }
      } catch (error) {
        // eslint-disable-line no-console
      }
    }
    loadStorageSettings()
  }, [getApiUrl])

  // Handle storage settings change
  const handleStorageChange = event => {
    const { name, value } = event.target
    setStorageSettings(prev => ({
      ...prev,
      [name]: value,
    }))
  }

  // Save storage settings - add better error handling
  const saveStorageSettings = async () => {
    setLoading(true)
    setStorageStatus({ message: t('SettingsDialog.Saving'), type: 'info' })

    try {
      const response = await fetch(getApiUrl('/storage/settings'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(storageSettings),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Failed to save settings')
      }

      if (result.status === 'ok') {
        setStorageStatus({
          message: t('SettingsDialog.StorageSettingsSaved'),
          type: 'success',
        })
      } else {
        setStorageStatus({
          message: t('SettingsDialog.SaveError') + (result.error || 'Unknown error'),
          type: 'error',
        })
      }
    } catch (error) {
      setStorageStatus({
        message: t('SettingsDialog.SaveError') + error.message,
        type: 'error',
      })
    } finally {
      setLoading(false)
    }
  }

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
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={EnableDebug} onChange={inputForm} id='EnableDebug' color='secondary' />}
          label={t('SettingsDialog.EnableDebug')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.EnableDebugHint')}</FormHelperText>
      </FormGroup>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={ResponsiveMode} onChange={inputForm} id='ResponsiveMode' color='secondary' />}
          label={t('SettingsDialog.ResponsiveMode')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.ResponsiveModeHint')}</FormHelperText>
      </FormGroup>
      <br />
      <FormGroup style={{ marginBottom: '20px' }}>
        <InputLabel htmlFor='RetrackersMode'>{t('SettingsDialog.RetrackersMode')}</InputLabel>
        <Select
          native
          type='number'
          id='RetrackersMode'
          name='RetrackersMode'
          value={RetrackersMode}
          onChange={inputForm}
          variant='outlined'
          margin='dense'
        >
          <option value={0}>{t('SettingsDialog.DontAddRetrackers')}</option>
          <option value={1}>{t('SettingsDialog.AddRetrackers')}</option>
          <option value={2}>{t('SettingsDialog.RemoveRetrackers')}</option>
          <option value={3}>{t('SettingsDialog.ReplaceRetrackers')}</option>
        </Select>
        <FormHelperText style={{ marginTop: '8px' }}>{t('SettingsDialog.RetrackersModeHint')}</FormHelperText>
      </FormGroup>
      {/* DLNA Section */}
      <SettingSectionLabel style={{ marginTop: '20px' }}>{t('DLNA')}</SettingSectionLabel>
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
      {/* HTTPS Section */}
      <SettingSectionLabel style={{ marginTop: '20px' }}>{t('HTTPS')}</SettingSectionLabel>
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
      {/* TorrFS */}
      <SettingSectionLabel style={{ marginTop: '20px' }}>{t('TorrFS')}</SettingSectionLabel>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={ShowFSActiveTorr} onChange={inputForm} id='ShowFSActiveTorr' color='secondary' />}
          label={t('SettingsDialog.ShowFSActiveTorr')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.ShowFSActiveTorrHint')}</FormHelperText>
      </FormGroup>
      {/* Storage Settings Section */}
      <Box mt={4} mb={2}>
        <SettingSectionLabel>{t('SettingsDialog.StorageConfiguration')}</SettingSectionLabel>

        <FormGroup>
          <InputLabel htmlFor='settings'>{t('SettingsDialog.SettingsStorage')}</InputLabel>
          <Select
            id='settings'
            name='settings'
            value={storageSettings.settings || 'json'}
            onChange={handleStorageChange}
            variant='outlined'
            margin='dense'
          >
            <MenuItem value='json'>{t('SettingsDialog.JsonFile')} (settings.json)</MenuItem>
            <MenuItem value='bbolt'>{t('SettingsDialog.BBoltDatabase')} (config.db)</MenuItem>
          </Select>
          <FormHelperText style={{ marginTop: '8px' }}>{t('SettingsDialog.SettingsStorageHint')}</FormHelperText>
        </FormGroup>

        <FormGroup style={{ marginTop: '16px' }}>
          <InputLabel htmlFor='viewed'>{t('SettingsDialog.ViewedHistoryStorage')}</InputLabel>
          <Select
            id='viewed'
            name='viewed'
            value={storageSettings.viewed || 'bbolt'}
            onChange={handleStorageChange}
            variant='outlined'
            margin='dense'
          >
            <MenuItem value='bbolt'>{t('SettingsDialog.BBoltDatabase')} (config.db)</MenuItem>
            <MenuItem value='json'>{t('SettingsDialog.JsonFile')} (viewed.json)</MenuItem>
          </Select>
          <FormHelperText style={{ marginTop: '8px' }}>{t('SettingsDialog.ViewedStorageHint')}</FormHelperText>
        </FormGroup>

        <Box mt={2} mb={2}>
          <Button
            variant='contained'
            color='primary'
            onClick={saveStorageSettings}
            disabled={loading}
            startIcon={loading ? <CircularProgress size={20} /> : null}
          >
            {t('SettingsDialog.SaveStorageSettings')}
          </Button>
        </Box>

        {storageStatus.message && (
          <StatusMessage severity={storageStatus.type}>
            <span>{storageStatus.message}</span>
            <Button onClick={() => setStorageStatus({ message: '', type: '' })} size='small'>
              ×
            </Button>
          </StatusMessage>
        )}
      </Box>
      {/* ProxyP2P */}
      <SettingSectionLabel style={{ marginTop: '20px' }}>{t('Proxy')}</SettingSectionLabel>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={EnableProxy} onChange={inputForm} id='EnableProxy' color='secondary' />}
          label={t('SettingsDialog.EnableProxy')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.EnableProxyHint')}</FormHelperText>
      </FormGroup>
      {/* Proxy hosts */}
      <TextField
        onChange={e => {
          setProxyHostsText(e.target.value)
        }}
        onBlur={e => {
          const inputValue = e.target.value.trim()
          const hostsArray =
            inputValue === ''
              ? []
              : inputValue
                  .split(',')
                  .map(s => s.trim())
                  .filter(s => s !== '')

          inputForm({
            target: {
              id: 'ProxyHosts',
              value: hostsArray,
            },
          })
        }}
        margin='normal'
        id='ProxyHosts'
        label={t('SettingsDialog.ProxyHosts')}
        helperText={t('SettingsDialog.ProxyHostsHint')}
        value={proxyHostsText}
        type='text'
        variant='outlined'
        fullWidth
      />
    </SecondarySettingsContent>
  )
}
