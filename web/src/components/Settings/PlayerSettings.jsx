import { useTranslation } from 'react-i18next'
import {
  Button,
  ButtonGroup,
  FormControl,
  FormControlLabel,
  FormHelperText,
  FormLabel,
  MenuItem,
  Switch,
  TextField,
} from '@material-ui/core'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function PlayerSettings({
  isAssStyled,
  setIsAssStyled,
  assRenderer,
  setAssRenderer,
  assFonts,
  setAssFonts,
}) {
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

      {isAssStyled && (
        <>
          <TextField
            select
            fullWidth
            margin='normal'
            variant='outlined'
            label={t('SettingsDialog.AssRenderer')}
            value={assRenderer}
            onChange={e => setAssRenderer(e.target.value)}
            helperText={t('SettingsDialog.AssRendererHint')}
          >
            <MenuItem value='auto'>{t('SettingsDialog.AssRendererAuto')}</MenuItem>
            <MenuItem value='full'>{t('SettingsDialog.AssRendererFull')}</MenuItem>
            <MenuItem value='simple'>{t('SettingsDialog.AssRendererSimple')}</MenuItem>
          </TextField>

          <FormControlLabel
            control={<Switch checked={assFonts} onChange={e => setAssFonts(e.target.checked)} color='secondary' />}
            label={t('SettingsDialog.AssFonts')}
            labelPlacement='start'
          />
          <div>
            <small>{t('SettingsDialog.AssFontsHint')}</small>
          </div>
        </>
      )}
    </SecondarySettingsContent>
  )
}
