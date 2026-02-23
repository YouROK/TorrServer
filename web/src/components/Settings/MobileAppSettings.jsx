import { FormControlLabel, FormGroup, FormHelperText, Switch, Link } from '@material-ui/core'
import { isMacOS, isAppleDevice, isDesktop } from 'utils/Utils'
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
  const isDesktopPlatform = isDesktop()

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
        {isDesktopPlatform && (
          <FormHelperText margin='none'>
            {t('SettingsDialog.UseVLCDesktopHintPrefix')}{' '}
            <Link
              href='https://github.com/northsea4/vlc-protocol'
              target='_blank'
              rel='noopener noreferrer'
              color='secondary'
            >
              vlc-protocol-handler
            </Link>
          </FormHelperText>
        )}
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
