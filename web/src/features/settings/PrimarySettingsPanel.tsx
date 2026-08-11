import { useEffect, useRef, useState } from 'react'
import { Database, Gauge, HardDrive } from 'lucide-react'
import { Input, Label, Slider, TextField } from '@heroui/react'
import { useTranslation } from 'react-i18next'

import type { BTSets } from 'shared/api/types'

import { SettingSwitch } from './SettingSwitch'
import SettingsSection from './SettingsSection'

const CACHE_SIZE_MIN_MB = 16
/** Slider ceiling only — the number field may go higher (e.g. 10240 = 10 GB). */
const CACHE_SIZE_SLIDER_MAX_MB = 2048

function clampCacheSizeMb(value: number): number {
  if (!Number.isFinite(value)) return CACHE_SIZE_MIN_MB
  return Math.max(CACHE_SIZE_MIN_MB, Math.round(value))
}

function CacheSizeField({ value, onChange }: { value: number; onChange: (mb: number) => void }) {
  const [text, setText] = useState(String(value))
  const editingRef = useRef(false)

  useEffect(() => {
    if (!editingRef.current) setText(String(value))
  }, [value])

  const commit = (raw: string) => {
    const next = clampCacheSizeMb(raw.trim() === '' ? Number.NaN : Number(raw))
    onChange(next)
    setText(String(next))
  }

  return (
    <TextField
      value={text}
      onChange={next => {
        setText(next)
        if (next.trim() === '') return
        const parsed = Number(next)
        if (Number.isFinite(parsed) && parsed >= CACHE_SIZE_MIN_MB) {
          onChange(Math.round(parsed))
        }
      }}
      onFocus={() => {
        editingRef.current = true
      }}
      onBlur={() => {
        editingRef.current = false
        commit(text)
      }}
      className='mt-2 max-w-[140px]'
    >
      <Input type='number' min={CACHE_SIZE_MIN_MB} step={16} />
    </TextField>
  )
}

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
  const preloadPct = settings.PreloadCache ?? 50
  const preloadSizeMb = Math.round((cacheSizeMb * Math.max(0, preloadPct)) / 100)

  return (
    <div className='space-y-6'>
      <SettingsSection icon={<Database />} title={t('SettingsDialog.SectionCache')}>
        <div>
          <p className='mb-2 text-sm text-muted'>
            {t('SettingsDialog.CacheSize')}: <span className='font-medium text-foreground'>{cacheSizeMb}</span>{' '}
            {t('MB')}
          </p>
          <Slider
            value={Math.min(cacheSizeMb, CACHE_SIZE_SLIDER_MAX_MB)}
            minValue={CACHE_SIZE_MIN_MB}
            maxValue={CACHE_SIZE_SLIDER_MAX_MB}
            step={16}
            onChange={value => onCacheSizeMb(Number(value))}
          >
            <Slider.Track>
              <Slider.Fill className='bg-accent' />
              <Slider.Thumb />
            </Slider.Track>
          </Slider>
          <CacheSizeField value={cacheSizeMb} onChange={onCacheSizeMb} />
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
            <span className='font-medium text-foreground'>
              {preloadPct}% ({preloadSizeMb} {t('MB')})
            </span>
          </p>
          <Slider
            value={preloadPct}
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
