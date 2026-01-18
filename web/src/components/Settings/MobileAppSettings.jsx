import { FormControlLabel, FormGroup, FormHelperText, Switch } from '@material-ui/core'
import { isMacOS, isAppleDevice } from 'utils/Utils'
import { useTranslation } from 'react-i18next'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function MobileAppSettings({
  isVlcUsed,
  setIsVlcUsed,
  isInfuseUsed,
  setIsInfuseUsed,
  isIinaUsed,
  setIsIinaUsed,
}) {
  const { t } = useTranslation()
  const isMac = isMacOS()
  const isApple = isAppleDevice()

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
        {isApple && (
          <>
            <FormControlLabel
              control={
                <Switch checked={isInfuseUsed} onChange={() => setIsInfuseUsed(prev => !prev)} color='secondary' />
              }
              label={t('SettingsDialog.UseInfuse')}
              labelPlacement='start'
            />
            <FormHelperText margin='none'>{t('SettingsDialog.UseInfuseHint')}</FormHelperText>
          </>
        )}
        {isMac && (
          <>
            <FormControlLabel
              control={<Switch checked={isIinaUsed} onChange={() => setIsIinaUsed(prev => !prev)} color='secondary' />}
              label={t('SettingsDialog.UseIINA')}
              labelPlacement='start'
            />
            <FormHelperText margin='none'>{t('SettingsDialog.UseIINAHint')}</FormHelperText>
          </>
        )}
      </FormGroup>
    </SecondarySettingsContent>
  )
}
