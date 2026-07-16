import { useTranslation } from 'react-i18next'
import {
  Box,
  Button,
  CircularProgress,
  FormControlLabel,
  FormGroup,
  FormHelperText,
  IconButton,
  MenuItem,
  Switch,
  TextField,
} from '@material-ui/core'
import CloseIcon from '@material-ui/icons/Close'
import { useEffect, useMemo, useState } from 'react'
import { useQueryClient } from 'react-query'
import { gstSettingsHost } from 'utils/Hosts'
import { GST_RUNTIME_QUERY_KEY } from 'utils/GStreamer'

import {
  Divider,
  GstRuntimeStatusItem,
  GstRuntimeStatusList,
  GstSettingsContent,
  GstSubsectionLabel,
  SettingSectionLabel,
  SettingsStatusMessage,
} from './style'

const GST_MIN_VERSION = 1.22

const parseDecimalInput = value => {
  const normalized = String(value).trim().replace(',', '.')
  if (normalized === '' || normalized === '.') {
    return null
  }
  const num = Number(normalized)
  return Number.isFinite(num) ? num : null
}

const formatDecimalInput = value => {
  if (value == null || !Number.isFinite(value)) {
    return ''
  }
  return String(value)
}

const emptyConfig = {
  GSTVersion: GST_MIN_VERSION,
  GSTPath: '',
  Source: 'stream',
  MaxTasks: 0,
  InactiveMinutes: 5,
  AACBitrateKbps: 256,
  AACChannels: 0,
  AACSamplerate: 0,
  SegmentSeconds: 6,
  SegmentDiff: 20,
  Subtitles: true,
  TranscodeH264: false,
  TranscodeH265: false,
  TranscodeAV1: false,
  TranscodeVP9: false,
  TranscodeVP8: false,
  TranscodeAVI: false,
  HDRToSDR: false,
  HardwareAcceleration: true,
  UseGPU: true,
  X264Ultrafast: false,
  VideoBitrate: 10000,
}

const componentStatusKind = component => {
  if (!component) return 'missing'
  if (component.works) return 'ok'
  if (component.available) return 'warn'
  if (component.found) return 'warn'
  return 'missing'
}

export default function GStreamerSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [gstreamerSettings, setGstreamerSettings] = useState(emptyConfig)
  const [defaults, setDefaults] = useState(emptyConfig)
  const [status, setStatus] = useState({ message: '', type: '' })
  const [loading, setLoading] = useState(false)
  const [echoStatus, setEchoStatus] = useState(null)
  const [gstVersionText, setGstVersionText] = useState(formatDecimalInput(emptyConfig.GSTVersion))

  const gstSettingsUrl = gstSettingsHost()
  const gstEchoUrl = useMemo(() => {
    const base = gstSettingsUrl.replace(/\/gst\/settings$/, '')
    return `${base}/gst/echo`
  }, [gstSettingsUrl])

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const [settingsResponse, echoResponse] = await Promise.all([fetch(gstSettingsUrl), fetch(gstEchoUrl)])

        if (settingsResponse.ok) {
          const data = await settingsResponse.json()
          if (!data.built_in) {
            return
          }
          const config = data.config || emptyConfig
          setGstreamerSettings(config)
          setDefaults(data.defaults || emptyConfig)
          setGstVersionText(formatDecimalInput(config.GSTVersion))
        }

        if (echoResponse.ok) {
          setEchoStatus(await echoResponse.json())
        }
      } catch (error) {
        // ignore load errors
      }
    }

    loadSettings()
  }, [gstSettingsUrl, gstEchoUrl])

  const updateField = (field, value) => {
    setGstreamerSettings(prev => ({ ...prev, [field]: value }))
  }

  const normalizeVersion = version => {
    const parsed = parseDecimalInput(version)
    if (parsed === null) {
      return GST_MIN_VERSION
    }
    return Math.max(parsed, GST_MIN_VERSION)
  }

  const saveSettings = async () => {
    setLoading(true)
    setStatus({ message: t('SettingsDialog.Saving'), type: 'info' })

    try {
      const gstVersion = normalizeVersion(gstVersionText)
      setGstVersionText(formatDecimalInput(gstVersion))
      const config = {
        ...gstreamerSettings,
        GSTVersion: gstVersion,
      }
      const response = await fetch(gstSettingsUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: 'set', config }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Failed to save settings')
      }

      setGstreamerSettings(config)
      queryClient.invalidateQueries(GST_RUNTIME_QUERY_KEY)
      setStatus({
        message: t('GStreamer.SettingsSaved'),
        type: 'success',
      })
    } catch (error) {
      setStatus({
        message: t('SettingsDialog.SaveError') + error.message,
        type: 'error',
      })
    } finally {
      setLoading(false)
    }
  }

  const resetToDefaults = async () => {
    setLoading(true)
    setStatus({ message: t('SettingsDialog.Saving'), type: 'info' })

    try {
      const response = await fetch(gstSettingsUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: 'def' }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.error || 'Failed to reset settings')
      }

      const settingsResponse = await fetch(gstSettingsUrl)
      if (settingsResponse.ok) {
        const data = await settingsResponse.json()
        const config = data.config || defaults
        setGstreamerSettings(config)
        setDefaults(data.defaults || defaults)
        setGstVersionText(formatDecimalInput(config.GSTVersion))
      }
      queryClient.invalidateQueries(GST_RUNTIME_QUERY_KEY)
      setStatus({
        message: t('GStreamer.SettingsSaved'),
        type: 'success',
      })
    } catch (error) {
      setStatus({
        message: t('SettingsDialog.SaveError') + error.message,
        type: 'error',
      })
    } finally {
      setLoading(false)
    }
  }

  const renderRuntimeStatus = (label, component) => {
    const kind = componentStatusKind(component)
    const state =
      kind === 'ok'
        ? t('GStreamer.StatusWorks')
        : kind === 'warn'
        ? t('GStreamer.StatusAvailable')
        : t('GStreamer.StatusMissing')

    return (
      <GstRuntimeStatusItem key={label} ok={kind === 'ok'} warn={kind === 'warn'}>
        <div className='gst-status-row'>
          <span className='gst-status-label'>{label}</span>
          <span className='gst-status-value'>{state}</span>
        </div>
        {component?.error ? <div className='gst-status-error'>{component.error}</div> : null}
      </GstRuntimeStatusItem>
    )
  }

  return (
    <GstSettingsContent>
      <SettingSectionLabel>{t('GStreamer.Settings')}</SettingSectionLabel>

      {echoStatus && (
        <GstRuntimeStatusList>
          {renderRuntimeStatus(t('GStreamer.Runtime'), echoStatus.gstreamer)}
          {renderRuntimeStatus(t('GStreamer.Discoverer'), echoStatus.gst_discoverer)}
        </GstRuntimeStatusList>
      )}

      <GstSubsectionLabel>{t('GStreamer.SectionGeneral')}</GstSubsectionLabel>

      <TextField
        label={t('GStreamer.Version')}
        type='text'
        value={gstVersionText}
        onChange={e => {
          const raw = e.target.value
          if (!/^[\d.,]*$/.test(raw)) {
            return
          }
          setGstVersionText(raw)
        }}
        onBlur={() => {
          const normalized = normalizeVersion(gstVersionText)
          updateField('GSTVersion', normalized)
          setGstVersionText(formatDecimalInput(normalized))
        }}
        margin='normal'
        helperText={t('GStreamer.VersionHint')}
        variant='outlined'
        fullWidth
        inputProps={{ inputMode: 'decimal', pattern: '[0-9]*[.,]?[0-9]*' }}
      />

      <TextField
        label={t('GStreamer.Path')}
        value={gstreamerSettings.GSTPath || ''}
        onChange={e => updateField('GSTPath', e.target.value)}
        margin='normal'
        helperText={t('GStreamer.PathHint')}
        variant='outlined'
        fullWidth
      />

      <TextField
        select
        id='gstreamer-source'
        label={t('GStreamer.Source')}
        value={gstreamerSettings.Source || 'stream'}
        onChange={e => updateField('Source', e.target.value)}
        margin='normal'
        helperText={t('GStreamer.SourceHint')}
        variant='outlined'
        fullWidth
      >
        <MenuItem value='stream'>{t('GStreamer.SourceStream')}</MenuItem>
        <MenuItem value='play'>{t('GStreamer.SourcePlay')}</MenuItem>
      </TextField>

      <Divider />

      <GstSubsectionLabel>{t('GStreamer.SectionPipeline')}</GstSubsectionLabel>

      <TextField
        label={t('GStreamer.MaxTasks')}
        type='number'
        value={gstreamerSettings.MaxTasks ?? 0}
        onChange={e => updateField('MaxTasks', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.MaxTasksHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 0 }}
      />

      <TextField
        label={t('GStreamer.InactiveMinutes')}
        type='number'
        value={gstreamerSettings.InactiveMinutes}
        onChange={e => updateField('InactiveMinutes', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.InactiveMinutesHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 1 }}
      />

      <TextField
        label={t('GStreamer.SegmentSeconds')}
        type='number'
        value={gstreamerSettings.SegmentSeconds}
        onChange={e => updateField('SegmentSeconds', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.SegmentSecondsHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 1 }}
      />

      <TextField
        label={t('GStreamer.SegmentDiff')}
        type='number'
        value={gstreamerSettings.SegmentDiff ?? 20}
        onChange={e => updateField('SegmentDiff', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.SegmentDiffHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 0 }}
      />

      <Divider />

      <GstSubsectionLabel>{t('GStreamer.SectionAudio')}</GstSubsectionLabel>

      <TextField
        label={t('GStreamer.AACBitrateKbps')}
        type='number'
        value={gstreamerSettings.AACBitrateKbps}
        onChange={e => updateField('AACBitrateKbps', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.AACBitrateKbpsHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 32 }}
      />

      <TextField
        label={t('GStreamer.AACChannels')}
        type='number'
        value={gstreamerSettings.AACChannels ?? 0}
        onChange={e => updateField('AACChannels', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.AACChannelsHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 0 }}
      />

      <TextField
        label={t('GStreamer.AACSamplerate')}
        type='number'
        value={gstreamerSettings.AACSamplerate ?? 0}
        onChange={e => updateField('AACSamplerate', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.AACSamplerateHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 0 }}
      />

      <Divider />

      <GstSubsectionLabel>{t('GStreamer.SectionTranscoding')}</GstSubsectionLabel>

      <TextField
        label={t('GStreamer.VideoBitrate')}
        type='number'
        value={gstreamerSettings.VideoBitrate}
        onChange={e => updateField('VideoBitrate', Number(e.target.value))}
        margin='normal'
        helperText={t('GStreamer.VideoBitrateHint')}
        variant='outlined'
        fullWidth
        inputProps={{ min: 100 }}
      />

      <FormHelperText style={{ marginBottom: 8 }}>{t('GStreamer.TranscodeHint')}</FormHelperText>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.TranscodeH264)}
              onChange={e => updateField('TranscodeH264', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.TranscodeH264')}
          labelPlacement='start'
        />
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.TranscodeH265)}
              onChange={e => updateField('TranscodeH265', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.TranscodeH265')}
          labelPlacement='start'
        />
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.TranscodeAV1)}
              onChange={e => updateField('TranscodeAV1', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.TranscodeAV1')}
          labelPlacement='start'
        />
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.TranscodeVP9)}
              onChange={e => updateField('TranscodeVP9', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.TranscodeVP9')}
          labelPlacement='start'
        />
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.TranscodeVP8)}
              onChange={e => updateField('TranscodeVP8', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.TranscodeVP8')}
          labelPlacement='start'
        />
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.TranscodeAVI)}
              onChange={e => updateField('TranscodeAVI', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.TranscodeAVI')}
          labelPlacement='start'
        />
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.HDRToSDR)}
              onChange={e => updateField('HDRToSDR', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.HDRToSDR')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('GStreamer.HDRToSDRHint')}</FormHelperText>
      </FormGroup>

      <Divider />

      <GstSubsectionLabel>{t('GStreamer.SectionAdvanced')}</GstSubsectionLabel>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.Subtitles)}
              onChange={e => updateField('Subtitles', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.Subtitles')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('GStreamer.SubtitlesHint')}</FormHelperText>
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.UseGPU)}
              onChange={e => updateField('UseGPU', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.UseGPU')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('GStreamer.UseGPUHint')}</FormHelperText>
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.HardwareAcceleration)}
              onChange={e => updateField('HardwareAcceleration', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.HardwareAcceleration')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('GStreamer.HardwareAccelerationHint')}</FormHelperText>
      </FormGroup>

      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={Boolean(gstreamerSettings.X264Ultrafast)}
              onChange={e => updateField('X264Ultrafast', e.target.checked)}
              color='secondary'
            />
          }
          label={t('GStreamer.X264Ultrafast')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('GStreamer.X264UltrafastHint')}</FormHelperText>
      </FormGroup>

      <Box mt={3} mb={2} display='flex' flexWrap='wrap' style={{ gap: 10 }}>
        <Button
          type='button'
          variant='contained'
          color='primary'
          onClick={saveSettings}
          disabled={loading}
          startIcon={loading ? <CircularProgress size={20} /> : null}
        >
          {t('GStreamer.SaveSettings')}
        </Button>
        <Button type='button' variant='outlined' color='secondary' onClick={resetToDefaults} disabled={loading}>
          {t('SettingsDialog.ResetToDefault')}
        </Button>
      </Box>

      {status.message && (
        <SettingsStatusMessage severity={status.type}>
          <span>{status.message}</span>
          <IconButton
            type='button'
            aria-label={t('Close')}
            onClick={() => setStatus({ message: '', type: '' })}
            size='small'
          >
            <CloseIcon fontSize='small' />
          </IconButton>
        </SettingsStatusMessage>
      )}
    </GstSettingsContent>
  )
}
