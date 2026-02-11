import { FormControlLabel, FormGroup, FormHelperText, Switch, FormControl, InputLabel, Select, MenuItem } from '@material-ui/core'
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
  isPlayerUsed,
  setIsPlayerUsed,
  preferredPlayer,
  setPreferredPlayer,
}) {
  const { t } = useTranslation()
  const isMac = isMacOS()
  const isApple = isAppleDevice()

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('SettingsDialog.MobileAppSettings')}</SettingSectionLabel>
      <FormGroup>
        <FormControlLabel
          control={<Switch checked={isPlayerUsed} onChange={() => setIsPlayerUsed(prev => !prev)} color='secondary' />}
          label={t('SettingsDialog.UseVLC')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('SettingsDialog.UseVLCHint')}</FormHelperText>

        <FormGroup style={{ marginBottom: '20px', marginTop: 8 }}>
          <InputLabel htmlFor='PreferredPlayer'>{t('SettingsDialog.PreferredPlayer')}</InputLabel>
          <Select
            native
            id='PreferredPlayer'
            value={preferredPlayer}
            onChange={e => setPreferredPlayer(e.target.value)}
            variant='outlined'
            margin='dense'
            disabled={!isPlayerUsed}
          >
            <option value='vlc'>VLC</option>
            <option value='potplayer'>PotPlayer</option>
          </Select>
          <FormHelperText style={{ marginTop: '8px' }}>{t('SettingsDialog.PreferredPlayerHint')}</FormHelperText>
        </FormGroup>
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
