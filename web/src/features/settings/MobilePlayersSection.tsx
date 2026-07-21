import { Clapperboard } from 'lucide-react'
import { Description, Link, ListBox, Select } from '@heroui/react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { useLocalBoolPref, useLocalJsonPref } from 'shared/hooks/useLocalPref'
import { isAppleDevice, isDesktop, isMacOS, detectApplePlatform } from 'shared/lib/platform'
import {
  availablePosterPlayActions,
  coercePosterPlayAction,
  defaultPosterPlayAction,
  POSTER_PLAY_ACTION_KEY,
  type PosterPlayAction,
} from 'shared/lib/posterPlay'

import { SettingSwitch } from './SettingSwitch'
import SettingsSection from './SettingsSection'

const PLAYER_KEYS = {
  vlc: 'isVlcUsed',
  infuse: 'isInfuseUsed',
  senPlayer: 'isSenPlayerUsed',
  iina: 'isIinaUsed',
} as const

/** External-player quick-open toggles — persisted as local prefs, gated by detected platform. */
export default function MobilePlayersSection() {
  const { t } = useTranslation()
  const isMac = isMacOS()
  const isApple = isAppleDevice()
  const isDesktopPlatform = isDesktop()
  const isIOS = detectApplePlatform().isIOS

  const [isVlcUsed, setIsVlcUsed] = useLocalBoolPref(PLAYER_KEYS.vlc)
  const [isInfuseUsed, setIsInfuseUsed] = useLocalBoolPref(PLAYER_KEYS.infuse)
  const [isSenPlayerUsed, setIsSenPlayerUsed] = useLocalBoolPref(PLAYER_KEYS.senPlayer)
  const [isIinaUsed, setIsIinaUsed] = useLocalBoolPref(PLAYER_KEYS.iina)

  const playerFlags = useMemo(
    () => ({ isVlcUsed, isInfuseUsed, isSenPlayerUsed, isIinaUsed, isApple, isMac }),
    [isVlcUsed, isInfuseUsed, isSenPlayerUsed, isIinaUsed, isApple, isMac],
  )

  const [posterPlayStored, setPosterPlayStored] = useLocalJsonPref<PosterPlayAction>(
    POSTER_PLAY_ACTION_KEY,
    defaultPosterPlayAction(isIOS),
  )
  const posterPlayAction = coercePosterPlayAction(posterPlayStored, playerFlags, isIOS)
  const posterPlayOptions = availablePosterPlayActions(playerFlags)

  const posterPlayLabel = (action: PosterPlayAction): string => {
    switch (action) {
      case 'builtin':
        return t('SettingsDialog.PosterPlayBuiltin')
      case 'copyLink':
        return t('CopyLink')
      case 'vlc':
        return t('VLC')
      case 'infuse':
        return t('Infuse')
      case 'senPlayer':
        return t('SenPlayer')
      case 'iina':
        return t('IINA')
      default:
        return action
    }
  }

  return (
    <SettingsSection
      icon={<Clapperboard />}
      title={t('SettingsDialog.MobileAppSettings')}
      description={t('SettingsDialog.MobileAppInstantHint')}
    >
      <div className='space-y-1'>
        <SettingSwitch
          id={PLAYER_KEYS.vlc}
          label={t('SettingsDialog.UseVLC')}
          helper={t('SettingsDialog.UseVLCHint')}
          checked={isVlcUsed}
          onChange={(_id, checked) => setIsVlcUsed(checked)}
        />
        {isDesktopPlatform ? (
          <Description className='mt-1'>
            {t('SettingsDialog.UseVLCDesktopHintPrefix')}{' '}
            <Link href='https://github.com/northsea4/vlc-protocol' target='_blank' rel='noopener noreferrer'>
              vlc-protocol-handler
            </Link>
          </Description>
        ) : null}
      </div>

      {isApple ? (
        <>
          <SettingSwitch
            id={PLAYER_KEYS.infuse}
            label={t('SettingsDialog.UseInfuse')}
            helper={t('SettingsDialog.UseInfuseHint')}
            checked={isInfuseUsed}
            onChange={(_id, checked) => setIsInfuseUsed(checked)}
          />
          <SettingSwitch
            id={PLAYER_KEYS.senPlayer}
            label={t('SettingsDialog.UseSenPlayer')}
            helper={t('SettingsDialog.UseSenPlayerHint')}
            checked={isSenPlayerUsed}
            onChange={(_id, checked) => setIsSenPlayerUsed(checked)}
          />
        </>
      ) : null}

      {isMac ? (
        <SettingSwitch
          id={PLAYER_KEYS.iina}
          label={t('SettingsDialog.UseIINA')}
          helper={t('SettingsDialog.UseIINAHint')}
          checked={isIinaUsed}
          onChange={(_id, checked) => setIsIinaUsed(checked)}
        />
      ) : null}

      <div className='space-y-1.5 pt-2'>
        <p className='block text-sm font-medium leading-snug text-foreground'>{t('SettingsDialog.PosterPlayAction')}</p>
        <p className='block text-sm leading-relaxed text-muted'>{t('SettingsDialog.PosterPlayActionHint')}</p>
        <Select
          selectedKey={posterPlayAction}
          onSelectionChange={key => {
            if (key == null) return
            setPosterPlayStored(String(key) as PosterPlayAction)
          }}
        >
          <Select.Trigger className='min-h-11 w-full'>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox>
              {posterPlayOptions.map(action => (
                <ListBox.Item key={action} id={action} textValue={posterPlayLabel(action)}>
                  {posterPlayLabel(action)}
                </ListBox.Item>
              ))}
            </ListBox>
          </Select.Popover>
        </Select>
        <p className='block text-sm leading-relaxed text-muted'>{t('SettingsDialog.TorrsProtocolHint')}</p>
      </div>
    </SettingsSection>
  )
}
