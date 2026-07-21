import { useMemo, useState } from 'react'
import { Check, Maximize2, Moon, Palette, Search, Sun, SunMoon } from 'lucide-react'
import { Input, Label, TextField, ToggleButton, ToggleButtonGroup } from '@heroui/react'
import { useTranslation } from 'react-i18next'

import { DIALOGS_FULLSCREEN_PREF } from 'shared/hooks/useDialogFullScreen'
import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import { THEME_PALETTE_SWATCHES } from 'shared/theme/paletteSwatches'
import { THEME_MODES, THEME_PALETTE_IDS, useThemePreference, type ThemePalette } from 'shared/theme/useThemePreference'
import { iconMenu } from 'shared/ui/iconProps'

import { SettingSwitch } from './SettingSwitch'
import SettingsSection from './SettingsSection'

function paletteLabelKey(id: ThemePalette): string {
  return `ThemePalette${id.charAt(0).toUpperCase()}${id.slice(1)}`
}

/** Client appearance — brightness + named palette; persists in localStorage immediately. */
export default function AppearanceSettingsPanel() {
  const { t } = useTranslation()
  const { preference, setPreference, palette, setPalette } = useThemePreference()
  const [dialogsFullScreen, setDialogsFullScreen] = useLocalBoolPref(DIALOGS_FULLSCREEN_PREF, false)
  const [query, setQuery] = useState('')

  const labels = useMemo(
    () => Object.fromEntries(THEME_PALETTE_IDS.map(id => [id, t(paletteLabelKey(id))])) as Record<ThemePalette, string>,
    [t],
  )

  const filtered = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return THEME_PALETTE_IDS
    return THEME_PALETTE_IDS.filter(id => {
      const name = labels[id].toLocaleLowerCase()
      return name.includes(q) || id.includes(q)
    })
  }, [labels, query])

  return (
    <div className='space-y-6'>
      <SettingsSection
        icon={<Palette />}
        title={t('SettingsDialog.SectionAppearance')}
        description={t('SettingsDialog.AppearanceInstantHint')}
      >
        <div>
          <Label className='mb-2 block text-sm'>{t('Theme')}</Label>
          <ToggleButtonGroup
            selectionMode='single'
            selectedKeys={[preference]}
            className='flex flex-wrap gap-1'
            aria-label={t('Theme')}
          >
            <ToggleButton
              id={THEME_MODES.LIGHT}
              onPress={() => setPreference(THEME_MODES.LIGHT)}
              className='min-h-11 gap-2'
            >
              <Sun {...iconMenu} aria-hidden />
              {t('ThemeLight')}
            </ToggleButton>
            <ToggleButton
              id={THEME_MODES.DARK}
              onPress={() => setPreference(THEME_MODES.DARK)}
              className='min-h-11 gap-2'
            >
              <Moon {...iconMenu} aria-hidden />
              {t('ThemeDark')}
            </ToggleButton>
            <ToggleButton
              id={THEME_MODES.AUTO}
              onPress={() => setPreference(THEME_MODES.AUTO)}
              className='min-h-11 gap-2'
            >
              <SunMoon {...iconMenu} aria-hidden />
              {t('ThemeAuto')}
            </ToggleButton>
          </ToggleButtonGroup>
        </div>

        <div>
          <div className='mb-2 flex flex-wrap items-end justify-between gap-2'>
            <Label className='text-sm'>{t('ThemePalette')}</Label>
            <p className='text-xs text-muted'>
              {labels[palette]}
              {filtered.length !== THEME_PALETTE_IDS.length
                ? ` · ${filtered.length}/${THEME_PALETTE_IDS.length}`
                : null}
            </p>
          </div>

          <div className='relative mb-3'>
            <Search
              size={16}
              strokeWidth={1.75}
              className='pointer-events-none absolute top-1/2 left-3 z-10 -translate-y-1/2 text-muted'
              aria-hidden
            />
            <TextField value={query} onChange={setQuery} aria-label={t('SettingsDialog.ThemePaletteSearch')}>
              <Input type='search' placeholder={t('SettingsDialog.ThemePaletteSearch')} className='min-h-11 pl-9' />
            </TextField>
          </div>

          {filtered.length === 0 ? (
            <p className='py-6 text-center text-sm text-muted'>{t('SettingsDialog.ThemePaletteEmpty')}</p>
          ) : (
            <div className='grid grid-cols-2 gap-2 sm:grid-cols-3' role='listbox' aria-label={t('ThemePalette')}>
              {filtered.map(id => {
                const selected = id === palette
                const swatch = THEME_PALETTE_SWATCHES[id]
                return (
                  <button
                    key={id}
                    type='button'
                    role='option'
                    aria-selected={selected}
                    onClick={() => setPalette(id)}
                    className={`flex min-h-11 items-center gap-2.5 rounded-lg border px-2.5 py-2 text-left transition-colors ${
                      selected
                        ? 'border-accent bg-accent/10 ring-1 ring-accent/40'
                        : 'border-border bg-surface hover-fine:border-accent/40 hover-fine:bg-surface-tertiary'
                    }`}
                  >
                    <span
                      className='relative size-8 shrink-0 overflow-hidden rounded-md border border-black/10 shadow-sm'
                      aria-hidden
                      style={{
                        background: `linear-gradient(135deg, ${swatch.header} 0 52%, ${swatch.accent} 52% 100%)`,
                      }}
                    >
                      {selected ? (
                        <span className='absolute inset-0 grid place-items-center bg-black/25'>
                          <Check size={14} strokeWidth={2.5} className='text-white' />
                        </span>
                      ) : null}
                    </span>
                    <span className='min-w-0 flex-1 truncate text-sm font-medium'>{labels[id]}</span>
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </SettingsSection>

      <SettingsSection
        icon={<Maximize2 />}
        title={t('SettingsDialog.SectionDialogs')}
        description={t('SettingsDialog.DialogsInstantHint')}
      >
        <SettingSwitch
          id={DIALOGS_FULLSCREEN_PREF}
          label={t('SettingsDialog.DialogsFullScreen')}
          helper={t('SettingsDialog.DialogsFullScreenHint')}
          checked={dialogsFullScreen}
          onChange={(_id, checked) => setDialogsFullScreen(checked)}
        />
      </SettingsSection>
    </div>
  )
}
