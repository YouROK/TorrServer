import axios from 'axios'
import {
  Button,
  Checkbox,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  Snackbar,
  Switch,
  TextField,
} from '@material-ui/core'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { playbackDeviceConfigsHost, playbackDevicesHost, playbackSettingsHost } from 'utils/Hosts'
import { PLAYBACK_CLIENT_MODES, PLAYBACK_ROUTING_MODES, usePlayback } from 'playback/PlaybackContext'

import { Divider, SecondarySettingsContent, SettingSectionLabel } from './style'

const defaultServerSettings = {
  enabled: false,
  routing_mode: PLAYBACK_ROUTING_MODES.LOCAL,
  primary_device_id: '',
}

const emptyForm = {
  id: '',
  name: '',
  endpoint: '',
  stream_base_url: '',
  token: '',
  clear_token: false,
  fullscreen: false,
  has_token: false,
}

const PlaybackSettings = () => {
  const { t } = useTranslation()
  const { mode, refreshPlaybackConfig, setMode } = usePlayback()
  const [serverSettings, setServerSettings] = useState(defaultServerSettings)
  const [configs, setConfigs] = useState([])
  const [form, setForm] = useState(emptyForm)
  const [isLoading, setIsLoading] = useState(true)
  const [isSavingSettings, setIsSavingSettings] = useState(false)
  const [isSavingDevice, setIsSavingDevice] = useState(false)
  const [testingId, setTestingId] = useState('')
  const [deleteCandidate, setDeleteCandidate] = useState(null)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const loadConfigs = useCallback(async () => {
    setIsLoading(true)
    try {
      const { data } = await axios.get(playbackDeviceConfigsHost())
      if (Array.isArray(data)) {
        setServerSettings(defaultServerSettings)
        setConfigs(data)
      } else {
        setServerSettings({
          enabled: Boolean(data?.enabled),
          routing_mode: data?.routing_mode || PLAYBACK_ROUTING_MODES.LOCAL,
          primary_device_id: data?.primary_device_id || '',
        })
        setConfigs(Array.isArray(data?.devices) ? data.devices : [])
      }
    } catch (_) {
      setError(t('Playback.LoadFailed'))
    } finally {
      setIsLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadConfigs()
  }, [loadConfigs])

  const updateForm =
    key =>
    ({ target: { value, checked, type } }) => {
      setForm(current => ({ ...current, [key]: type === 'checkbox' ? checked : value }))
    }

  const updateServerSetting =
    key =>
    ({ target: { value, checked, type } }) => {
      setServerSettings(current => ({ ...current, [key]: type === 'checkbox' ? checked : value }))
    }

  const editDevice = device => {
    setForm({
      id: device.id,
      name: device.name,
      endpoint: device.endpoint,
      stream_base_url: device.stream_base_url || '',
      token: '',
      clear_token: false,
      fullscreen: Boolean(device.fullscreen),
      has_token: Boolean(device.has_token),
    })
  }

  const resetForm = () => setForm(emptyForm)

  const saveRoutingSettings = async () => {
    setIsSavingSettings(true)
    try {
      await axios.post(playbackSettingsHost(), serverSettings)
      setMessage(t('Playback.SettingsSaved'))
      await Promise.all([loadConfigs(), refreshPlaybackConfig()])
    } catch (requestError) {
      setError(requestError.response?.data?.error || t('Playback.SettingsSaveFailed'))
    } finally {
      setIsSavingSettings(false)
    }
  }

  const saveDevice = async () => {
    setIsSavingDevice(true)
    try {
      await axios.post(playbackDevicesHost(), {
        id: form.id || undefined,
        name: form.name,
        endpoint: form.endpoint,
        stream_base_url: form.stream_base_url,
        token: form.token,
        clear_token: form.clear_token,
        fullscreen: form.fullscreen,
      })
      setMessage(t('Playback.DeviceSaved'))
      resetForm()
      await Promise.all([loadConfigs(), refreshPlaybackConfig()])
    } catch (requestError) {
      setError(requestError.response?.data?.error || t('Playback.SaveFailed'))
    } finally {
      setIsSavingDevice(false)
    }
  }

  const deleteDevice = async () => {
    if (!deleteCandidate) return
    try {
      await axios.delete(`${playbackDevicesHost()}/${encodeURIComponent(deleteCandidate.id)}`)
      if (form.id === deleteCandidate.id) resetForm()
      setMessage(t('Playback.DeviceDeleted'))
      setDeleteCandidate(null)
      await Promise.all([loadConfigs(), refreshPlaybackConfig()])
    } catch (_) {
      setError(t('Playback.DeleteFailed'))
    }
  }

  const testDevice = async id => {
    setTestingId(id)
    try {
      await axios.post(`${playbackDevicesHost()}/${encodeURIComponent(id)}/test`)
      setMessage(t('Playback.TestSuccessful'))
    } catch (_) {
      setError(t('Playback.TestFailed'))
    } finally {
      setTestingId('')
    }
  }

  const primaryMode = serverSettings.routing_mode === PLAYBACK_ROUTING_MODES.PRIMARY
  const perBrowserMode = serverSettings.routing_mode === PLAYBACK_ROUTING_MODES.PER_BROWSER

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('Playback.RemotePlayback')}</SettingSectionLabel>
      <FormControlLabel
        control={
          <Switch color='secondary' checked={serverSettings.enabled} onChange={updateServerSetting('enabled')} />
        }
        label={t('Playback.EnableRemotePlayback')}
        labelPlacement='start'
      />
      <FormHelperText>{t('Playback.EnableRemotePlaybackHint')}</FormHelperText>

      <FormControl fullWidth variant='outlined' margin='normal'>
        <InputLabel id='playback-routing-mode-label'>{t('Playback.RoutingMode')}</InputLabel>
        <Select
          labelId='playback-routing-mode-label'
          label={t('Playback.RoutingMode')}
          value={serverSettings.routing_mode}
          disabled={!serverSettings.enabled}
          onChange={updateServerSetting('routing_mode')}
        >
          <MenuItem value={PLAYBACK_ROUTING_MODES.LOCAL}>{t('Playback.RoutingLocal')}</MenuItem>
          <MenuItem value={PLAYBACK_ROUTING_MODES.PRIMARY}>{t('Playback.RoutingPrimary')}</MenuItem>
          <MenuItem value={PLAYBACK_ROUTING_MODES.PER_BROWSER}>{t('Playback.RoutingPerBrowser')}</MenuItem>
        </Select>
        <FormHelperText>{t('Playback.RoutingModeHint')}</FormHelperText>
      </FormControl>

      {serverSettings.enabled && primaryMode && (
        <FormControl fullWidth variant='outlined' margin='normal'>
          <InputLabel id='primary-playback-device-label'>{t('Playback.PrimaryDevice')}</InputLabel>
          <Select
            labelId='primary-playback-device-label'
            label={t('Playback.PrimaryDevice')}
            value={serverSettings.primary_device_id}
            onChange={updateServerSetting('primary_device_id')}
          >
            {configs.length === 0 && (
              <MenuItem value='' disabled>
                {t('Playback.NoRemoteDevices')}
              </MenuItem>
            )}
            {configs.map(device => (
              <MenuItem key={device.id} value={device.id}>
                {device.name}
              </MenuItem>
            ))}
          </Select>
          <FormHelperText>{t('Playback.PrimaryDeviceHint')}</FormHelperText>
        </FormControl>
      )}

      {serverSettings.enabled && perBrowserMode && (
        <FormControl fullWidth variant='outlined' margin='normal'>
          <InputLabel id='playback-client-mode-label'>{t('Playback.BrowserRole')}</InputLabel>
          <Select
            labelId='playback-client-mode-label'
            label={t('Playback.BrowserRole')}
            value={mode}
            onChange={({ target: { value } }) => setMode(value)}
          >
            <MenuItem value={PLAYBACK_CLIENT_MODES.CONTROL_AND_PLAYBACK}>{t('Playback.ControlAndPlayback')}</MenuItem>
            <MenuItem value={PLAYBACK_CLIENT_MODES.CONTROL_ONLY}>{t('Playback.ControlOnly')}</MenuItem>
          </Select>
          <FormHelperText>{t('Playback.BrowserRoleHint')}</FormHelperText>
        </FormControl>
      )}

      <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 12 }}>
        <Button
          variant='contained'
          color='secondary'
          disabled={isSavingSettings || (serverSettings.enabled && primaryMode && !serverSettings.primary_device_id)}
          onClick={saveRoutingSettings}
        >
          {isSavingSettings ? <CircularProgress size={20} color='inherit' /> : t('Playback.SaveRouting')}
        </Button>
      </div>

      <Divider />

      <SettingSectionLabel>{t('Playback.RemoteDevices')}</SettingSectionLabel>
      <FormHelperText>{t('Playback.RemoteDevicesHint')}</FormHelperText>

      {isLoading ? (
        <div style={{ display: 'grid', placeItems: 'center', minHeight: 80 }}>
          <CircularProgress color='secondary' size={28} />
        </div>
      ) : (
        <div style={{ display: 'grid', gap: 10, marginTop: 16 }}>
          {configs.length === 0 && <FormHelperText>{t('Playback.NoRemoteDevices')}</FormHelperText>}
          {configs.map(device => (
            <div
              key={device.id}
              style={{
                border: '1px solid rgba(127,127,127,0.35)',
                borderRadius: 6,
                padding: 12,
                display: 'grid',
                gridTemplateColumns: 'minmax(0, 1fr) auto',
                gap: 12,
                alignItems: 'center',
              }}
            >
              <div style={{ minWidth: 0 }}>
                <div style={{ fontWeight: 600 }}>{device.name}</div>
                <div style={{ fontSize: 12, opacity: 0.75, overflowWrap: 'anywhere' }}>{device.endpoint}</div>
                <div style={{ fontSize: 12, opacity: 0.75 }}>
                  {device.fullscreen ? t('Playback.FullscreenEnabled') : t('Playback.FullscreenDisabled')}
                </div>
              </div>
              <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
                <Button size='small' color='secondary' onClick={() => editDevice(device)}>
                  {t('Edit')}
                </Button>
                <Button
                  size='small'
                  color='secondary'
                  disabled={testingId === device.id}
                  onClick={() => testDevice(device.id)}
                >
                  {testingId === device.id ? <CircularProgress size={16} color='inherit' /> : t('Playback.Test')}
                </Button>
                <Button size='small' color='secondary' onClick={() => setDeleteCandidate(device)}>
                  {t('Delete')}
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Divider />

      <SettingSectionLabel>{form.id ? t('Playback.EditDevice') : t('Playback.AddDevice')}</SettingSectionLabel>
      <div style={{ display: 'grid', gap: 14, marginTop: 16 }}>
        <TextField
          required
          variant='outlined'
          label={t('Playback.DeviceName')}
          value={form.name}
          onChange={updateForm('name')}
        />
        <TextField
          required
          type='url'
          variant='outlined'
          label={t('Playback.AgentURL')}
          placeholder='http://192.168.1.20:8092'
          value={form.endpoint}
          onChange={updateForm('endpoint')}
          helperText={t('Playback.AgentURLHint')}
        />
        <TextField
          type='url'
          variant='outlined'
          label={t('Playback.StreamBaseURL')}
          placeholder='http://192.168.1.10:8090'
          value={form.stream_base_url}
          onChange={updateForm('stream_base_url')}
          helperText={t('Playback.StreamBaseURLHint')}
        />
        <TextField
          type='password'
          variant='outlined'
          label={t('Playback.AgentToken')}
          value={form.token}
          onChange={updateForm('token')}
          helperText={form.has_token ? t('Playback.TokenConfiguredHint') : t('Playback.AgentTokenHint')}
        />
        <FormControlLabel
          control={<Checkbox color='secondary' checked={form.fullscreen} onChange={updateForm('fullscreen')} />}
          label={t('Playback.OpenFullscreen')}
        />
        <FormHelperText>{t('Playback.OpenFullscreenHint')}</FormHelperText>
        {form.id && form.has_token && (
          <FormControlLabel
            control={<Checkbox color='secondary' checked={form.clear_token} onChange={updateForm('clear_token')} />}
            label={t('Playback.ClearToken')}
          />
        )}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, flexWrap: 'wrap' }}>
          {form.id && (
            <Button variant='outlined' color='secondary' onClick={resetForm}>
              {t('Cancel')}
            </Button>
          )}
          <Button
            variant='contained'
            color='secondary'
            disabled={isSavingDevice || !form.name.trim() || !form.endpoint.trim()}
            onClick={saveDevice}
          >
            {isSavingDevice ? <CircularProgress size={20} color='inherit' /> : t('Save')}
          </Button>
        </div>
      </div>

      <Dialog open={Boolean(deleteCandidate)} onClose={() => setDeleteCandidate(null)}>
        <DialogTitle>{t('Playback.DeleteDevice')}</DialogTitle>
        <DialogContent>{t('Playback.DeleteConfirm', { name: deleteCandidate?.name || '' })}</DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteCandidate(null)} color='secondary'>
            {t('Cancel')}
          </Button>
          <Button onClick={deleteDevice} color='secondary' variant='contained'>
            {t('Delete')}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar open={Boolean(message)} autoHideDuration={2500} onClose={() => setMessage('')} message={message} />
      <Snackbar open={Boolean(error)} autoHideDuration={3500} onClose={() => setError('')} message={error} />
    </SecondarySettingsContent>
  )
}

export default PlaybackSettings
