import { Lock, Plug, Radio, ShieldCheck } from 'lucide-react'
import { Description, Input, Label, ListBox, Select, TextField } from '@heroui/react'
import { useTranslation } from 'react-i18next'

import type { BTSets } from 'shared/api/types'

import { SettingSwitch } from './SettingSwitch'
import SettingsSection from './SettingsSection'

export interface NetworkSettingsPanelProps {
  settings: BTSets
  boolChecked: (key: keyof BTSets) => boolean
  onUpdate: <K extends keyof BTSets>(key: K, value: BTSets[K]) => void
  onBoolSwitch: (id: string, checked: boolean) => void
}

const NUMBER_FIELD_DEFAULTS: Partial<Record<keyof BTSets, number>> = {
  ConnectionsLimit: 25,
  TorrentDisconnectTimeout: 30,
}

const NUMBER_FIELDS: readonly (keyof BTSets)[] = [
  'PeersListenPort',
  'ConnectionsLimit',
  'DownloadRateLimit',
  'UploadRateLimit',
  'TorrentDisconnectTimeout',
  'SslPort',
]

/** Protocol toggles, retracker mode, and connection/port tuning ("PRO mode"). */
export default function NetworkSettingsPanel({
  settings,
  boolChecked,
  onUpdate,
  onBoolSwitch,
}: NetworkSettingsPanelProps) {
  const { t } = useTranslation()

  const numberFieldLabel = (key: keyof BTSets): string => {
    switch (key) {
      case 'PeersListenPort':
        return t('SettingsDialog.PeersListenPort')
      case 'ConnectionsLimit':
        return t('SettingsDialog.ConnectionsLimit')
      case 'DownloadRateLimit':
        return t('SettingsDialog.DownloadRateLimit')
      case 'UploadRateLimit':
        return t('SettingsDialog.UploadRateLimit')
      case 'TorrentDisconnectTimeout':
        return t('SettingsDialog.TorrentDisconnectTimeout')
      case 'SslPort':
        return t('SettingsDialog.SslPort')
      default:
        return String(key)
    }
  }

  const numberFieldHint = (key: keyof BTSets): string | undefined => {
    switch (key) {
      case 'PeersListenPort':
        return t('SettingsDialog.PeersListenPortHint')
      case 'ConnectionsLimit':
        return t('SettingsDialog.ConnectionsLimitHint')
      case 'DownloadRateLimit':
      case 'UploadRateLimit':
        return t('SettingsDialog.RateLimitHint')
      case 'TorrentDisconnectTimeout':
        return t('Seconds')
      case 'SslPort':
        return t('SettingsDialog.SslPortHint')
      default:
        return undefined
    }
  }

  return (
    <div className='space-y-6'>
      <p className='text-xs font-semibold tracking-wide text-muted uppercase'>{t('SettingsDialog.ProMode')}</p>

      <SettingsSection icon={<Radio />} title={t('SettingsDialog.SectionProtocols')}>
        <div className='grid gap-x-8 gap-y-4 sm:grid-cols-2'>
          <SettingSwitch
            id='DisableTCP'
            label='TCP'
            helper={t('SettingsDialog.DisableTCPHint')}
            checked={boolChecked('DisableTCP')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='DisableUTP'
            label='μTP'
            helper={t('SettingsDialog.DisableUTPHint')}
            checked={boolChecked('DisableUTP')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='DisableUPNP'
            label='UPnP'
            helper={t('SettingsDialog.DisableUPNPHint')}
            checked={boolChecked('DisableUPNP')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='DisableDHT'
            label={t('SettingsDialog.DHT')}
            helper={t('SettingsDialog.DisableDHTHint')}
            checked={boolChecked('DisableDHT')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='DisablePEX'
            label='PEX'
            helper={t('SettingsDialog.DisablePEXHint')}
            checked={boolChecked('DisablePEX')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='EnableIPv6'
            label='IPv6'
            helper={t('SettingsDialog.EnableIPv6Hint')}
            checked={boolChecked('EnableIPv6')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='EnableLPD'
            label='LPD'
            helper={t('SettingsDialog.EnableLPDHint')}
            checked={boolChecked('EnableLPD')}
            onChange={onBoolSwitch}
          />
          <SettingSwitch
            id='LPDIPv6'
            label='LPD IPv6'
            helper={t('SettingsDialog.EnableLPDIPv6Hint')}
            checked={boolChecked('LPDIPv6')}
            onChange={onBoolSwitch}
          />
        </div>
      </SettingsSection>

      <SettingsSection icon={<Lock />} title={t('SettingsDialog.SectionUploadEncryption')}>
        <SettingSwitch
          id='DisableUpload'
          label={t('SettingsDialog.Upload')}
          helper={t('SettingsDialog.UploadHint')}
          checked={boolChecked('DisableUpload')}
          onChange={onBoolSwitch}
        />
        <SettingSwitch
          id='ForceEncrypt'
          label={t('SettingsDialog.ForceEncrypt')}
          helper={t('SettingsDialog.ForceEncryptHint')}
          checked={boolChecked('ForceEncrypt')}
          onChange={onBoolSwitch}
        />
      </SettingsSection>

      <SettingsSection icon={<ShieldCheck />} title={t('SettingsDialog.RetrackersMode')}>
        <Select
          selectedKey={settings.RetrackersMode == null ? '1' : String(settings.RetrackersMode)}
          onSelectionChange={key => onUpdate('RetrackersMode', Number(key))}
        >
          <Select.Trigger>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox>
              <ListBox.Item id='0'>{t('SettingsDialog.DontAddRetrackers')}</ListBox.Item>
              <ListBox.Item id='1'>{t('SettingsDialog.AddRetrackers')}</ListBox.Item>
              <ListBox.Item id='2'>{t('SettingsDialog.RemoveRetrackers')}</ListBox.Item>
              <ListBox.Item id='3'>{t('SettingsDialog.ReplaceRetrackers')}</ListBox.Item>
            </ListBox>
          </Select.Popover>
        </Select>
        <p className='text-sm text-muted'>{t('SettingsDialog.RetrackersModeHint')}</p>
      </SettingsSection>

      <SettingsSection icon={<Plug />} title={t('SettingsDialog.SectionConnection')}>
        <div className='grid gap-4 sm:grid-cols-2'>
          {NUMBER_FIELDS.map(key => (
            <TextField
              key={key}
              value={String(settings[key] ?? NUMBER_FIELD_DEFAULTS[key] ?? 0)}
              onChange={value => onUpdate(key, Number(value))}
            >
              <Label>{numberFieldLabel(key)}</Label>
              <Input type='number' />
              {numberFieldHint(key) ? <Description>{numberFieldHint(key)}</Description> : null}
            </TextField>
          ))}
        </div>
      </SettingsSection>

      <SettingsSection title={t('SettingsDialog.SectionSslTls')}>
        <TextField value={settings.SslCert || ''} onChange={value => onUpdate('SslCert', value)}>
          <Label>{t('SettingsDialog.SslCert')}</Label>
          <Input placeholder='/etc/ssl/cert.pem' />
          <Description>{t('SettingsDialog.SslCertHint')}</Description>
        </TextField>
        <TextField value={settings.SslKey || ''} onChange={value => onUpdate('SslKey', value)}>
          <Label>{t('SettingsDialog.SslKey')}</Label>
          <Input placeholder='/etc/ssl/key.pem' />
          <Description>{t('SettingsDialog.SslKeyHint')}</Description>
        </TextField>
      </SettingsSection>
    </div>
  )
}
