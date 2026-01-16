import { useTranslation } from 'react-i18next'
import { FormGroup, FormHelperText, TextField } from '@material-ui/core'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function TMDBSettings({ settings, updateSettings }) {
  const { t } = useTranslation()
  const { TMDBSettings } = settings || {}
  const {
    APIKey = '',
    APIURL = 'https://api.themoviedb.org/3',
    ImageURL = 'https://image.tmdb.org',
    ImageURLRu = 'https://imagetmdb.com',
  } = TMDBSettings || {}

  const handleChange = (field, value) => {
    updateSettings({
      TMDBSettings: {
        ...TMDBSettings,
        [field]: value,
      },
    })
  }

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>{t('TMDB.Settings')}</SettingSectionLabel>
      <FormGroup>
        <TextField
          label={t('TMDB.APIKey')}
          value={APIKey}
          onChange={e => handleChange('APIKey', e.target.value)}
          placeholder='Enter your TMDB API key'
          variant='outlined'
          size='small'
          fullWidth
          style={{ marginBottom: 15 }}
        />
        <FormHelperText margin='none'>{t('TMDB.APIKeyHint')}</FormHelperText>
      </FormGroup>

      <FormGroup style={{ marginTop: 20 }}>
        <TextField
          label={t('TMDB.APIURL')}
          value={APIURL}
          onChange={e => handleChange('APIURL', e.target.value)}
          placeholder='https://api.themoviedb.org/3'
          variant='outlined'
          size='small'
          fullWidth
          style={{ marginBottom: 10 }}
        />
        <FormHelperText margin='none'>{t('TMDB.APIURLHint')}</FormHelperText>
      </FormGroup>

      <FormGroup style={{ marginTop: 20 }}>
        <TextField
          label={t('TMDB.ImageURL')}
          value={ImageURL}
          onChange={e => handleChange('ImageURL', e.target.value)}
          placeholder='https://image.tmdb.org'
          variant='outlined'
          size='small'
          fullWidth
          style={{ marginBottom: 10 }}
        />
        <FormHelperText margin='none'>{t('TMDB.ImageURLHint')}</FormHelperText>
      </FormGroup>

      <FormGroup style={{ marginTop: 20 }}>
        <TextField
          label={t('TMDB.ImageURLRu')}
          value={ImageURLRu}
          onChange={e => handleChange('ImageURLRu', e.target.value)}
          placeholder='https://imagetmdb.com'
          variant='outlined'
          size='small'
          fullWidth
          style={{ marginBottom: 10 }}
        />
        <FormHelperText margin='none'>{t('TMDB.ImageURLRuHint')}</FormHelperText>
      </FormGroup>
      <br />
    </SecondarySettingsContent>
  )
}
