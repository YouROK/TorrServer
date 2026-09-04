import { HardDrive } from 'lucide-react'
import { Alert, Label, ListBox, Select } from '@heroui/react'
import { useTranslation } from 'react-i18next'

import type { StorageBackend, StorageSettings } from 'shared/api/storage'

import SettingsSection from './SettingsSection'

export interface StorageSettingsPanelProps {
  backends: StorageSettings
  onBackendsChange: (next: StorageSettings) => void
}

/** Where BT settings + viewed history are persisted (JSON file vs. BBolt DB). */
export default function StorageSettingsPanel({ backends, onBackendsChange }: StorageSettingsPanelProps) {
  const { t } = useTranslation()

  const updateBackend = (key: keyof StorageSettings, value: StorageBackend) => {
    onBackendsChange({ ...backends, [key]: value })
  }

  return (
    <div className='space-y-6'>
      <SettingsSection
        icon={<HardDrive />}
        title={t('SettingsDialog.StorageSettings')}
        description={t('SettingsDialog.SettingsStorageHint')}
      >
        <Select
          selectedKey={backends.settings}
          onSelectionChange={key => updateBackend('settings', String(key) as StorageBackend)}
        >
          <Label>{t('SettingsDialog.SettingsStorage')}</Label>
          <Select.Trigger>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox>
              <ListBox.Item id='json'>{t('SettingsDialog.JsonFile')} (settings.json)</ListBox.Item>
              <ListBox.Item id='bbolt'>{t('SettingsDialog.BBoltDatabase')} (config.db)</ListBox.Item>
            </ListBox>
          </Select.Popover>
        </Select>

        <Select
          selectedKey={backends.viewed}
          onSelectionChange={key => updateBackend('viewed', String(key) as StorageBackend)}
        >
          <Label>{t('SettingsDialog.ViewedHistoryStorage')}</Label>
          <Select.Trigger>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox>
              <ListBox.Item id='bbolt'>{t('SettingsDialog.BBoltDatabase')} (config.db)</ListBox.Item>
              <ListBox.Item id='json'>{t('SettingsDialog.JsonFile')} (viewed.json)</ListBox.Item>
            </ListBox>
          </Select.Popover>
        </Select>
      </SettingsSection>

      {backends.viewedCount != null ? (
        <p className='text-sm text-muted'>
          {t('StorageViewedCount')}: <strong className='text-foreground'>{backends.viewedCount}</strong>
        </p>
      ) : null}

      <Alert status='accent'>
        <Alert.Description>{t('StorageRestartHint')}</Alert.Description>
      </Alert>
    </div>
  )
}
