import { Film } from 'lucide-react'
import { Description, Input, Label, TextField } from '@heroui/react'
import { useTranslation } from 'react-i18next'

import type { BTSets, SettingsUpdater, TMDBSettingsConfig } from 'shared/api/types'

import SettingsSection from './SettingsSection'

interface TMDBSettingsSectionProps {
  settings?: BTSets
  updateSettings: SettingsUpdater
}

type TMDBField = keyof Pick<TMDBSettingsConfig, 'APIKey' | 'APIURL' | 'ImageURL' | 'ImageURLRu'>

/** TMDB poster-search credentials — API key + overridable API/image base URLs. */
export default function TMDBSettingsSection({ settings, updateSettings }: TMDBSettingsSectionProps) {
  const { t } = useTranslation()
  const tmdb = settings?.TMDBSettings || {}

  const handleChange = (field: TMDBField, value: string) => {
    updateSettings({ TMDBSettings: { ...tmdb, [field]: value } })
  }

  const fields: Array<{ field: TMDBField; label: string; hint: string; placeholder: string }> = [
    {
      field: 'APIKey',
      label: t('TMDB.APIKey'),
      hint: t('TMDB.APIKeyHint'),
      placeholder: t('TMDB.APIKeyPlaceholder'),
    },
    {
      field: 'APIURL',
      label: t('TMDB.APIURL'),
      hint: t('TMDB.APIURLHint'),
      placeholder: t('TMDB.APIURLPlaceholder'),
    },
    {
      field: 'ImageURL',
      label: t('TMDB.ImageURL'),
      hint: t('TMDB.ImageURLHint'),
      placeholder: t('TMDB.ImageURLPlaceholder'),
    },
    {
      field: 'ImageURLRu',
      label: t('TMDB.ImageURLRu'),
      hint: t('TMDB.ImageURLRuHint'),
      placeholder: t('TMDB.ImageURLRuPlaceholder'),
    },
  ]

  return (
    <SettingsSection icon={<Film />} title={t('TMDB.Settings')}>
      {fields.map(({ field, label, hint, placeholder }) => (
        <TextField
          key={field}
          value={String(tmdb[field] ?? (field === 'APIKey' ? '' : placeholder))}
          onChange={value => handleChange(field, value)}
        >
          <Label>{label}</Label>
          <Input placeholder={placeholder} />
          <Description>{hint}</Description>
        </TextField>
      ))}
    </SettingsSection>
  )
}
