import { Database, Gauge, HardDrive } from 'lucide-react'
import { Input, Label, Slider, TextField } from '@heroui/react'
import { useTranslation } from 'react-i18next'

import type { BTSets } from 'shared/api/types'

import { SettingSwitch } from './SettingSwitch'
import SettingsSection from './SettingsSection'

export interface PrimarySettingsPanelProps {
  settings: BTSets
  cacheSizeMb: number
  onCacheSizeMb: (mb: number) => void
  onUpdate: <K extends keyof BTSets>(key: K, value: BTSets[K]) => void
  onBoolSwitch: (id: string, checked: boolean) => void
}

/** Cache size / readahead / preload / disk-storage controls — the server's primary tuning knobs. */
export default function PrimarySettingsPanel({
  settings,
  cacheSizeMb,
  onCacheSizeMb,
  onUpdate,
  onBoolSwitch,
}: PrimarySettingsPanelProps) {
  const { t } = useTranslation()

  return (
    <div className='space-y-6'>
      <SettingsSection icon={<Database />} title={t('SettingsDialog.SectionCache')}>
        <div>
          <p className='mb-2 text-sm text-muted'>
            {t('SettingsDialog.CacheSize')}: <span className='font-medium text-foreground'>{cacheSizeMb}</span>{' '}
            {t('MB')}
          </p>
          <Slider
            value={cacheSizeMb}
            minValue={16}
            maxValue={2048}
            step={16}
            onChange={value => onCacheSizeMb(Number(value))}
          >
            <Slider.Track>
              <Slider.Fill className='bg-accent' />
              <Slider.Thumb />
            </Slider.Track>
          </Slider>
          <TextField
            value={String(cacheSizeMb)}
            onChange={value => onCacheSizeMb(Math.max(16, Number(value) || 16))}
            className='mt-2 max-w-[140px]'
          >
            <Input type='number' min={16} step={16} />
          </TextField>
        </div>
      </SettingsSection>

      <SettingsSection icon={<Gauge />} title={t('SettingsDialog.SectionStreaming')}>
        <div>
          <p className='mb-2 text-sm text-muted'>
            {t('SettingsDialog.ReaderReadAHead')}:{' '}
            <span className='font-medium text-foreground'>{settings.ReaderReadAHead ?? 95}%</span>
          </p>
          <Slider
            value={settings.ReaderReadAHead ?? 95}
            minValue={5}
            maxValue={100}
            onChange={value => onUpdate('ReaderReadAHead', Number(value))}
          >
            <Slider.Track>
              <Slider.Fill className='bg-accent' />
              <Slider.Thumb />
            </Slider.Track>
          </Slider>
        </div>

        <div>
          <p className='mb-2 text-sm text-muted'>
            {t('SettingsDialog.PreloadCache')}:{' '}
            <span className='font-medium text-foreground'>{settings.PreloadCache ?? 50}%</span>
          </p>
          <Slider
            value={settings.PreloadCache ?? 50}
            minValue={0}
            maxValue={100}
            onChange={value => onUpdate('PreloadCache', Number(value))}
          >
            <Slider.Track>
              <Slider.Fill className='bg-accent' />
              <Slider.Thumb />
            </Slider.Track>
          </Slider>
        </div>
      </SettingsSection>

      <SettingsSection icon={<HardDrive />} title={t('SettingsDialog.SectionDiskCache')}>
        <SettingSwitch
          id='UseDisk'
          label={t('SettingsDialog.UseDisk')}
          helper={t('SettingsDialog.UseDiskDesc')}
          checked={Boolean(settings.UseDisk)}
          onChange={onBoolSwitch}
        />
        <TextField value={settings.TorrentsSavePath || ''} onChange={value => onUpdate('TorrentsSavePath', value)}>
          <Label>{t('SettingsDialog.TorrentsSavePath')}</Label>
          <Input placeholder='/data/torrents' />
        </TextField>
        <SettingSwitch
          id='RemoveCacheOnDrop'
          label={t('SettingsDialog.RemoveCacheOnDrop')}
          helper={t('SettingsDialog.RemoveCacheOnDropDesc')}
          checked={Boolean(settings.RemoveCacheOnDrop)}
          onChange={onBoolSwitch}
        />
      </SettingsSection>
    </div>
  )
}
