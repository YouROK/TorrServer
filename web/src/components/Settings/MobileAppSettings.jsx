import { FormControlLabel, FormGroup, FormHelperText, Switch } from '@material-ui/core'
import { useTranslation } from 'react-i18next'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function MobileAppSettings({ isVlcUsed, setIsVlcUsed, isInfuseUsed, setIsInfuseUsed }) {
  const { t } = useTranslation()

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('SettingsDialog.MobileAppSettings')}</SettingSectionLabel>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={isVlcUsed} onChange={() => setIsVlcUsed(prev => !prev)} color='secondary' />}
          label={t('SettingsDialog.UseVLC')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.UseVLCHint')}</FormHelperText>
        <FormControlLabel
          control={<Switch checked={isInfuseUsed} onChange={() => setIsInfuseUsed(prev => !prev)} color='secondary' />}
          label={t('SettingsDialog.UseInfuse')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.UseInfuseHint')}</FormHelperText>
      </FormGroup>
    </SecondarySettingsContent>
  )
}
