import { useEffect, useState } from 'react'
import axios from 'axios'
import { Box, FormControlLabel, FormGroup, FormHelperText, InputAdornment, Switch } from '@material-ui/core'
import TextField from '@material-ui/core/TextField'
import { useTranslation } from 'react-i18next'
import { torrentsHost } from 'utils/Hosts'

import { SettingSectionLabel } from './style'

const GIB = 1024 ** 3

const formatSize = bytes => {
  if (!(bytes > 0)) return '0 B'
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  let i = 0
  let v = bytes
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i += 1
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const toGbText = bytes => (bytes > 0 ? String(+(bytes / GIB).toFixed(2)) : '0')

// A size setting stored in bytes and edited in GiB. The text is kept locally, so typing
// "1." or "0.5" is not rewritten under the cursor.
function SizeField({ id, label, hint, bytes, inputForm }) {
  const { t } = useTranslation()
  const [text, setText] = useState(toGbText(bytes))

  useEffect(() => {
    const typed = Math.round((parseFloat(text) || 0) * GIB)
    if (typed !== (bytes || 0)) setText(toGbText(bytes))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bytes])

  const onChange = ({ target: { value } }) => {
    setText(value)
    const n = parseFloat(value)
    inputForm({ target: { type: 'number', id, value: Number.isFinite(n) && n > 0 ? Math.round(n * GIB) : 0 } })
  }

  return (
    <TextField
      onChange={onChange}
      margin='normal'
      id={id}
      label={label}
      helperText={hint}
      value={text}
      type='number'
      inputProps={{ min: 0, step: 0.5 }}
      InputProps={{ endAdornment: <InputAdornment position='end'>{t('GB')}</InputAdornment> }}
      variant='outlined'
      fullWidth
    />
  )
}

// "Watch Offline Mode" block of the Additional tab (same layout as "Storage Configuration").
// OfflineSavePath applies to torrents switched to the offline mode after the change;
// already downloaded ones stay where they are. Quotas are checked when an offline download
// starts or resumes; they are saved with the main Save button.
export default function OfflineSettings({ settings, inputForm }) {
  const { t } = useTranslation()
  const { OfflineSavePath, OfflineMaxSize, OfflineMinFreeSpace, OfflineRotation } = settings || {}
  const [quota, setQuota] = useState(null)

  useEffect(() => {
    let cancelled = false
    axios
      .post(torrentsHost(), { action: 'offline_quota' })
      .then(({ data }) => !cancelled && setQuota(data))
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <Box mt={4} mb={2}>
      <SettingSectionLabel>{t('SettingsDialog.WatchOfflineMode')}</SettingSectionLabel>

      <TextField
        onChange={inputForm}
        margin='normal'
        id='OfflineSavePath'
        label={t('SettingsDialog.OfflineSavePath')}
        helperText={t('SettingsDialog.OfflineSavePathHint')}
        value={OfflineSavePath || ''}
        type='text'
        variant='outlined'
        fullWidth
      />

      <SizeField
        id='OfflineMaxSize'
        label={t('SettingsDialog.OfflineMaxSize')}
        hint={t('SettingsDialog.OfflineMaxSizeHint')}
        bytes={OfflineMaxSize}
        inputForm={inputForm}
      />
      <SizeField
        id='OfflineMinFreeSpace'
        label={t('SettingsDialog.OfflineMinFreeSpace')}
        hint={t('SettingsDialog.OfflineMinFreeSpaceHint')}
        bytes={OfflineMinFreeSpace}
        inputForm={inputForm}
      />
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={!!OfflineRotation} onChange={inputForm} id='OfflineRotation' color='secondary' />}
          label={t('SettingsDialog.OfflineRotation')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.OfflineRotationHint')}</FormHelperText>
      </FormGroup>

      {quota && (
        <FormHelperText style={{ marginTop: 12 }}>
          {t('SettingsDialog.OfflineQuotaUsage', {
            used: formatSize(quota.used),
            limit: quota.max_size > 0 ? formatSize(quota.max_size) : t('SettingsDialog.OfflineUnlimited'),
            free: quota.free >= 0 ? formatSize(quota.free) : '?',
          })}
        </FormHelperText>
      )}
    </Box>
  )
}
