import { FormControlLabel, Switch } from '@material-ui/core'
import { useTranslation } from 'react-i18next'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function MobileAppSettings({ isVlcUsed, setIsVlcUsed }) {
  const { t } = useTranslation()

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('SettingsDialog.MobileAppSettings')}</SettingSectionLabel>

      <FormControlLabel
        control={<Switch checked={isVlcUsed} onChange={() => setIsVlcUsed(prev => !prev)} color='secondary' />}
        label={t('SettingsDialog.UseVLC')}
        labelPlacement='start'
      />
    </SecondarySettingsContent>
  )
}
