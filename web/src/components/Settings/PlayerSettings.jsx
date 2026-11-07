import { useTranslation } from 'react-i18next'
import { Button, ButtonGroup, FormControl, FormHelperText, FormLabel } from '@material-ui/core'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function PlayerSettings({ isAssStyled, setIsAssStyled }) {
  const { t } = useTranslation()

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('SettingsDialog.PlayerSettings')}</SettingSectionLabel>

      <FormControl fullWidth>
        <FormLabel style={{ marginBottom: 8 }}>{t('SettingsDialog.AssStyled')}</FormLabel>
        <ButtonGroup fullWidth color='secondary' aria-label={t('SettingsDialog.AssStyled')}>
          <Button
            variant={isAssStyled ? 'contained' : 'outlined'}
            aria-pressed={isAssStyled}
            onClick={() => setIsAssStyled(true)}
          >
            {t('SettingsDialog.AssFormatGraphic')}
          </Button>
          <Button
            variant={isAssStyled ? 'outlined' : 'contained'}
            aria-pressed={!isAssStyled}
            onClick={() => setIsAssStyled(false)}
          >
            {t('SettingsDialog.AssFormatText')}
          </Button>
        </ButtonGroup>
        <FormHelperText>{t('SettingsDialog.AssStyledHint')}</FormHelperText>
      </FormControl>
    </SecondarySettingsContent>
  )
}
