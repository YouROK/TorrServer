import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Box, Button, CircularProgress, FormHelperText, TextField } from '@material-ui/core'
import axios from 'axios'
import { wafHost } from 'utils/Hosts'

import { SecondarySettingsContent, SettingSectionLabel, SettingsStatusMessage, GstSubsectionLabel } from './style'

export default function WAFSettings({ onDirtyChange }) {
  const { t } = useTranslation()
  const [whitelist, setWhitelist] = useState('')
  const [blacklist, setBlacklist] = useState('')
  const [referers, setReferers] = useState('')
  const [warnings, setWarnings] = useState([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [loaded, setLoaded] = useState(false)
  const [loadError, setLoadError] = useState('')
  const [readOnly, setReadOnly] = useState(false)
  const [initial, setInitial] = useState({ whitelist: '', blacklist: '', referers: '' })
  const [status, setStatus] = useState({ message: '', type: '' })

  const applyResponse = useCallback(data => {
    const next = {
      whitelist: data.whitelist || '',
      blacklist: data.blacklist || '',
      referers: data.referers || '',
    }
    setWhitelist(next.whitelist)
    setBlacklist(next.blacklist)
    setReferers(next.referers)
    setInitial(next)
    setWarnings(data.warnings || [])
    setReadOnly(Boolean(data.read_only))
    setLoaded(true)
    setLoadError('')
  }, [])

  const loadSettings = useCallback(async () => {
    setLoading(true)
    setLoadError('')
    try {
      const { data } = await axios.get(wafHost())
      applyResponse(data)
    } catch (err) {
      setLoaded(false)
      setLoadError(err.response?.data?.error || err.message || '')
    } finally {
      setLoading(false)
    }
  }, [applyResponse])

  useEffect(() => {
    loadSettings()
  }, [loadSettings])

  const dirty =
    loaded && (whitelist !== initial.whitelist || blacklist !== initial.blacklist || referers !== initial.referers)

  useEffect(() => {
    onDirtyChange?.(dirty)
    return () => onDirtyChange?.(false)
  }, [dirty, onDirtyChange])

  useEffect(() => {
    const warnBeforeUnload = event => {
      if (!dirty) return
      event.preventDefault()
      // Browser compatibility requires assigning returnValue.
      // eslint-disable-next-line no-param-reassign
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', warnBeforeUnload)
    return () => window.removeEventListener('beforeunload', warnBeforeUnload)
  }, [dirty])

  const handleSave = async () => {
    setSaving(true)
    setStatus({ message: '', type: '' })
    try {
      const { data } = await axios.post(wafHost(), {
        whitelist,
        blacklist,
        referers,
      })
      applyResponse(data)
      setStatus({ message: t('WAF.Saved'), type: 'success' })
    } catch (err) {
      if (err.response?.status === 403) {
        setReadOnly(true)
      }
      setStatus({
        message:
          err.response?.status === 403
            ? t('WAF.ReadOnlyHint')
            : err.response?.data?.error || err.message || t('WAF.SaveFailed'),
        type: 'error',
      })
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <SecondarySettingsContent>
        <CircularProgress color='secondary' />
      </SecondarySettingsContent>
    )
  }

  if (!loaded) {
    return (
      <SecondarySettingsContent>
        <SettingsStatusMessage severity='error' role='alert' aria-live='assertive'>
          {loadError || t('WAF.LoadFailed')}
        </SettingsStatusMessage>
        <Box mt={2}>
          <Button variant='outlined' color='secondary' onClick={loadSettings}>
            {t('WAF.Retry')}
          </Button>
        </Box>
      </SecondarySettingsContent>
    )
  }

  const formatWarning = warning =>
    t(warning.line ? 'WAF.WarningMessage' : 'WAF.WarningMessageNoLine', {
      list: t(`WAF.Lists.${warning.list}`),
      line: warning.line,
      reason: t(`WAF.WarningCodes.${warning.code}`),
    })

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>
        {t('WAF.Title')}
        <small>{t('WAF.Subtitle')}</small>
      </SettingSectionLabel>

      <FormHelperText style={{ marginBottom: 16 }}>{t('WAF.StorageHint')}</FormHelperText>
      <FormHelperText style={{ marginBottom: 16 }}>{t('WAF.SeparateSaveHint')}</FormHelperText>

      {readOnly && (
        <SettingsStatusMessage severity='info' role='status' aria-live='polite'>
          {t('WAF.ReadOnlyHint')}
        </SettingsStatusMessage>
      )}

      <FormHelperText style={{ marginBottom: 16 }}>{t('WAF.IPRulesHint')}</FormHelperText>

      <TextField
        label={t('WAF.Whitelist')}
        helperText={t('WAF.WhitelistHint')}
        value={whitelist}
        onChange={e => setWhitelist(e.target.value)}
        disabled={readOnly}
        multiline
        minRows={4}
        fullWidth
        variant='outlined'
        margin='normal'
        InputProps={{ style: { fontFamily: 'monospace', fontSize: 13 } }}
      />

      <TextField
        label={t('WAF.Blacklist')}
        helperText={t('WAF.BlacklistHint')}
        value={blacklist}
        onChange={e => setBlacklist(e.target.value)}
        disabled={readOnly}
        multiline
        minRows={4}
        fullWidth
        variant='outlined'
        margin='normal'
        InputProps={{ style: { fontFamily: 'monospace', fontSize: 13 } }}
      />

      <GstSubsectionLabel>{t('WAF.ReferersSection')}</GstSubsectionLabel>
      <FormHelperText style={{ marginBottom: 8 }}>{t('WAF.ReferersHint')}</FormHelperText>

      <TextField
        label={t('WAF.Referers')}
        helperText={t('WAF.ReferersFieldHint')}
        value={referers}
        onChange={e => setReferers(e.target.value)}
        disabled={readOnly}
        multiline
        minRows={3}
        fullWidth
        variant='outlined'
        margin='normal'
        InputProps={{ style: { fontFamily: 'monospace', fontSize: 13 } }}
      />

      {warnings.length > 0 && (
        <SettingsStatusMessage severity='warning' role='alert' aria-live='polite'>
          <div>
            <div id='waf-warning-heading'>{t('WAF.Warnings')}</div>
            <ul aria-labelledby='waf-warning-heading' style={{ margin: '8px 0 0', paddingLeft: 20 }}>
              {warnings.map(warning => (
                <li key={`${warning.list}-${warning.line}-${warning.code}`}>{formatWarning(warning)}</li>
              ))}
            </ul>
          </div>
        </SettingsStatusMessage>
      )}

      {status.message && (
        <SettingsStatusMessage
          severity={status.type === 'success' ? 'success' : 'error'}
          role='alert'
          aria-live='polite'
        >
          {status.message}
        </SettingsStatusMessage>
      )}

      <Box mt={2} display='flex' justifyContent='flex-end'>
        <Button
          variant='contained'
          color='secondary'
          onClick={handleSave}
          disabled={saving || readOnly || !dirty}
          aria-busy={saving}
        >
          {saving ? <CircularProgress size={22} color='inherit' /> : t('WAF.SaveLists')}
        </Button>
      </Box>
    </SecondarySettingsContent>
  )
}
