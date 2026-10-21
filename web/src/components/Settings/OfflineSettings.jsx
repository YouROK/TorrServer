import { Box } from '@material-ui/core'
import TextField from '@material-ui/core/TextField'
import { useTranslation } from 'react-i18next'

import { SettingSectionLabel } from './style'

// "Watch Offline Mode" block of the Additional tab (same layout as "Storage Configuration").
// OfflineSavePath applies to torrents switched to the offline mode after the change;
// already downloaded ones stay where they are. It is saved with the main Save button.
export default function OfflineSettings({ settings, inputForm }) {
  const { t } = useTranslation()
  const { OfflineSavePath } = settings || {}

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
    </Box>
  )
}
