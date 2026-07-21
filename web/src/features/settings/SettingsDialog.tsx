import { Button, Description, Label, ListBox, Modal, Select, Spinner, Tabs, useMediaQuery } from '@heroui/react'
import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { useGSAP } from '@gsap/react'
import gsap from 'gsap'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Cog, Film, HardDrive, Palette, Rss, SlidersHorizontal, Smartphone, Wifi } from 'lucide-react'

import type { BTSets } from 'shared/api/types'
import { getSettings, setSettings, resetSettings, SETTINGS_QUERY_KEY } from 'shared/api/settings'
import { getGstSettings, setGstSettings, resetGstSettings } from 'shared/api/gst'
import {
  defaultStorageSettings,
  getStorageSettings,
  setStorageSettings,
  type StorageSettings,
} from 'shared/api/storage'
import defaultSettings from 'shared/settings/defaults'
import { GST_RUNTIME_QUERY_KEY, useGStreamerRuntime } from 'shared/lib/gstreamer'
import { notifySettingsChanged } from 'shared/lib/settingsEvents'
import { clearTMDBCache } from 'shared/lib/torrentHelpers'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { queryMax } from 'shared/theme/breakpoints'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SETTINGS } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

import AppearanceSettingsPanel from './AppearanceSettingsPanel'
import FeaturesSettingsPanel from './FeaturesSettingsPanel'
import GStreamerSettingsPanel, { emptyGstConfig, type GStreamerConfig } from './GStreamerSettingsPanel'
import MobilePlayersSection from './MobilePlayersSection'
import NetworkSettingsPanel from './NetworkSettingsPanel'
import PrimarySettingsPanel from './PrimarySettingsPanel'
import { DISABLE_SWITCH_IDS } from './SettingSwitch'
import StorageSettingsPanel from './StorageSettingsPanel'
import TMDBSettingsSection from './TMDBSettingsSection'
import TorznabSettingsPanel from './TorznabSettingsPanel'

export interface SettingsDialogProps {
  open: boolean
  onClose: () => void
  /** Deep-link into a specific tab when opened externally (e.g. an "enable external players" hint). */
  initialTab?: SettingsTab
}

type SettingsTab = 'primary' | 'network' | 'features' | 'storage' | 'appearance' | 'app' | 'gstreamer' | 'torznab'

/** Fades + slides in freshly-mounted tab content — `Tabs.Panel` only mounts the selected tab, so this
 *  re-runs on every switch without any extra wiring. */
function PanelFade({ children }: { children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null)
  useGSAP(() => {
    if (!ref.current) return
    const reduced = typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (reduced) {
      gsap.set(ref.current, { opacity: 1, y: 0 })
      return
    }
    gsap.fromTo(ref.current, { opacity: 0, y: 8 }, { opacity: 1, y: 0, duration: 0.22, ease: 'power2.out' })
  }, [])
  return <div ref={ref}>{children}</div>
}

/** Tabbed server settings dialog — full-screen on mobile, persists via BTSets + GST + storage-backend APIs.
 *
 * Mobile: horizontal Tabs.List was unreadable; a sticky section Select drives the same `tab` state.
 * Desktop: vertical Tabs. One global Save still applies to every section.
 */
export default function SettingsDialog({ open, onClose, initialTab }: SettingsDialogProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const queryClient = useQueryClient()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isFullScreenBreakpoint = useDialogFullScreen()
  const gstRuntime = useGStreamerRuntime()

  const [tab, setTab] = useState<SettingsTab>('primary')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [settings, setLocalSettings] = useState<BTSets>({ ...defaultSettings })
  const [cacheSizeMb, setCacheSizeMb] = useState(defaultSettings.CacheSize ?? 64)
  const [gstConfig, setGstConfig] = useState<GStreamerConfig>(emptyGstConfig())
  const [storageBackends, setStorageBackends] = useState<StorageSettings>(defaultStorageSettings())
  /** False when `/storage/settings` failed — Save still works for BTSets/GST; storage write is skipped. */
  const [storageLoadedOk, setStorageLoadedOk] = useState(false)

  const gstAvailable = Boolean(gstRuntime.built_in)

  const visibleTabs = useMemo(() => {
    const tabs: { id: SettingsTab; label: string; shortLabel: string; icon: ReactNode }[] = [
      {
        id: 'primary',
        label: t('SettingsDialog.Tabs.Main'),
        shortLabel: t('SettingsDialog.Tabs.Main'),
        icon: <SlidersHorizontal size={17} strokeWidth={1.75} />,
      },
      {
        id: 'network',
        label: t('SettingsDialog.Tabs.Network'),
        shortLabel: t('SettingsDialog.Tabs.Network'),
        icon: <Wifi size={17} strokeWidth={1.75} />,
      },
      {
        id: 'features',
        label: t('SettingsDialog.AdditionalSettings'),
        shortLabel: t('SettingsDialog.Tabs.AdvancedShort'),
        icon: <Cog size={17} strokeWidth={1.75} />,
      },
      {
        id: 'storage',
        label: t('SettingsDialog.StorageSettings'),
        shortLabel: t('SettingsDialog.Tabs.StorageShort'),
        icon: <HardDrive size={17} strokeWidth={1.75} />,
      },
      {
        id: 'appearance',
        label: t('SettingsDialog.Tabs.Appearance'),
        shortLabel: t('SettingsDialog.Tabs.Appearance'),
        icon: <Palette size={17} strokeWidth={1.75} />,
      },
      {
        id: 'app',
        label: t('SettingsDialog.Tabs.App'),
        shortLabel: t('SettingsDialog.Tabs.App'),
        icon: <Smartphone size={17} strokeWidth={1.75} />,
      },
    ]
    if (gstAvailable) {
      tabs.push({
        id: 'gstreamer',
        label: t('GStreamer.Tab'),
        shortLabel: t('GStreamer.Tab'),
        icon: <Film size={17} strokeWidth={1.75} />,
      })
    }
    tabs.push({
      id: 'torznab',
      label: t('Torznab.Tab'),
      shortLabel: t('Torznab.Tab'),
      icon: <Rss size={17} strokeWidth={1.75} />,
    })
    return tabs
  }, [gstAvailable, t])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- fall back if GST tab hidden on non-gst build
    if (!visibleTabs.some(item => item.id === tab)) setTab('primary')
  }, [tab, visibleTabs])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- deep-link / SettingsOpenWithTab
    if (open && initialTab) setTab(initialTab)
  }, [open, initialTab])

  const loadSettings = useCallback(
    async (signal?: AbortSignal) => {
      const data = await getSettings(signal)
      const loaded = { ...defaultSettings, ...data }
      const mb = Math.round((loaded.CacheSize ?? (defaultSettings.CacheSize ?? 64) * 1024 * 1024) / (1024 * 1024))
      const next = { ...loaded, CacheSize: mb }
      setLocalSettings(next)
      setCacheSizeMb(mb)
      // Keep play/resume TrackTimecode readers in sync even before Save.
      queryClient.setQueryData(SETTINGS_QUERY_KEY, { ...loaded, CacheSize: loaded.CacheSize })
    },
    [queryClient],
  )

  const loadGstConfig = useCallback(async (signal?: AbortSignal) => {
    try {
      const data = await getGstSettings(signal)
      if (!data.built_in) return
      setGstConfig({ ...emptyGstConfig(), ...(data.config || {}) })
    } catch {
      // GST is optional — not built in on this platform/build.
    }
  }, [])

  const loadStorageBackends = useCallback(
    async (signal?: AbortSignal) => {
      try {
        setStorageBackends(await getStorageSettings(signal))
        setStorageLoadedOk(true)
      } catch (err) {
        if ((err as Error).name === 'AbortError') return
        setStorageLoadedOk(false)
        toast?.showToast({ message: t('Error'), severity: 'error' })
      }
    },
    [t, toast],
  )

  useEffect(() => {
    if (!open) return undefined
    const ac = new AbortController()
    // eslint-disable-next-line react-hooks/set-state-in-effect -- load settings when dialog opens
    setLoading(true)
    setStorageLoadedOk(false)
    Promise.all([loadSettings(ac.signal), loadGstConfig(ac.signal), loadStorageBackends(ac.signal)])
      .catch(err => {
        if ((err as Error).name === 'AbortError') return
        toast?.showToast({ message: t('Error'), severity: 'error' })
      })
      .finally(() => {
        if (!ac.signal.aborted) setLoading(false)
      })
    return () => ac.abort()
  }, [open, loadSettings, loadGstConfig, loadStorageBackends, t, toast])

  const updateSetting = useCallback(<K extends keyof BTSets>(key: K, value: BTSets[K]) => {
    setLocalSettings(prev => ({ ...prev, [key]: value }))
  }, [])

  const updateSettingsPartial = useCallback((partial: Partial<BTSets>) => {
    setLocalSettings(prev => ({ ...prev, ...partial }))
  }, [])

  const handleBoolSwitch = useCallback(
    (id: string, checked: boolean) => {
      const key = id as keyof BTSets
      const value = DISABLE_SWITCH_IDS.has(id) ? !checked : checked
      updateSetting(key, value as BTSets[typeof key])
    },
    [updateSetting],
  )

  const boolChecked = useCallback(
    (key: keyof BTSets) => {
      const value = settings[key]
      if (DISABLE_SWITCH_IDS.has(String(key))) return !value
      return Boolean(value)
    },
    [settings],
  )

  const handleSave = async () => {
    if (saving) return
    setSaving(true)
    try {
      const sets: BTSets = {
        ...settings,
        CacheSize: cacheSizeMb * 1024 * 1024,
        ReaderReadAHead: settings.ReaderReadAHead ?? defaultSettings.ReaderReadAHead,
        PreloadCache: settings.PreloadCache ?? defaultSettings.PreloadCache,
      }
      await setSettings(sets)
      queryClient.setQueryData(SETTINGS_QUERY_KEY, sets)
      if (storageLoadedOk) {
        await setStorageSettings(storageBackends)
      }
      if (gstAvailable) {
        await setGstSettings(gstConfig)
        await queryClient.invalidateQueries({ queryKey: [GST_RUNTIME_QUERY_KEY] })
      }
      clearTMDBCache()
      notifySettingsChanged()
      await queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY })
      toast?.showToast({
        message: storageLoadedOk ? t('Saved') : t('SavedWithoutStorage'),
        severity: 'success',
      })
      onClose()
    } catch (err) {
      toast?.showToast({ message: (err as Error).message || t('Error'), severity: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const handleResetDefaults = async () => {
    if (saving) return
    setSaving(true)
    try {
      await resetSettings()
      const ac = new AbortController()
      await loadSettings(ac.signal)
      await loadGstConfig(ac.signal)
      clearTMDBCache()
      notifySettingsChanged()
      await queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY })
      toast?.showToast({ message: t('SettingsDialog.ResetToDefault'), severity: 'success' })
    } catch (err) {
      toast?.showToast({ message: (err as Error).message || t('Error'), severity: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const footerButtonClassName = isMobile ? 'min-h-11 px-4' : undefined
  const panelClassName = isMobile
    ? 'min-h-0 flex-1 overflow-y-auto pt-3'
    : 'ml-0 min-h-0 min-w-0 flex-1 overflow-y-auto'
  const tabsRootClassName = 'flex min-h-0 flex-1 gap-4 overflow-hidden'
  const tabsListClassName = 'sticky top-0 z-10 w-52 shrink-0 self-start bg-surface'

  /** Shared panel body for both the mobile Select path and desktop Tabs.Panel mounts. */
  const renderPanel = (id: SettingsTab) => {
    switch (id) {
      case 'primary':
        return (
          <PrimarySettingsPanel
            settings={settings}
            cacheSizeMb={cacheSizeMb}
            onCacheSizeMb={setCacheSizeMb}
            onUpdate={updateSetting}
            onBoolSwitch={handleBoolSwitch}
          />
        )
      case 'network':
        return (
          <NetworkSettingsPanel
            settings={settings}
            boolChecked={boolChecked}
            onUpdate={updateSetting}
            onBoolSwitch={handleBoolSwitch}
          />
        )
      case 'features':
        return (
          <FeaturesSettingsPanel
            settings={settings}
            boolChecked={boolChecked}
            onUpdate={updateSetting}
            onBoolSwitch={handleBoolSwitch}
          />
        )
      case 'storage':
        return (
          <>
            {!storageLoadedOk && !loading ? (
              <Description className='mb-3 text-warning'>{t('StorageLoadFailed')}</Description>
            ) : null}
            <StorageSettingsPanel backends={storageBackends} onBackendsChange={setStorageBackends} />
          </>
        )
      case 'appearance':
        return <AppearanceSettingsPanel />
      case 'app':
        return (
          <>
            <Description className='mb-4'>{t('SettingsDialog.AppTabHint')}</Description>
            <div className='space-y-6'>
              <TMDBSettingsSection settings={settings} updateSettings={updateSettingsPartial} />
              <MobilePlayersSection />
            </div>
          </>
        )
      case 'gstreamer':
        return (
          <>
            <GStreamerSettingsPanel config={gstConfig} onChange={setGstConfig} />
            <Button
              className='mt-4'
              variant='secondary'
              onPress={() => {
                void resetGstSettings()
                  .then(data => {
                    setGstConfig({ ...emptyGstConfig(), ...(data.config || data.defaults || {}) })
                    toast?.showToast({ message: t('SettingsDialog.ResetToDefault'), severity: 'success' })
                  })
                  .catch(() => toast?.showToast({ message: t('Error'), severity: 'error' }))
              }}
            >
              {t('SettingsDialog.ResetToDefault')}
            </Button>
          </>
        )
      case 'torznab':
        return (
          <TorznabSettingsPanel
            settings={settings}
            onUpdate={updateSetting}
            footerButtonClassName={footerButtonClassName}
          />
        )
      default:
        return null
    }
  }

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='lg'
      fullScreen={isFullScreenBreakpoint}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_SETTINGS}
    >
      <Modal.Header className='shrink-0'>
        <Modal.Heading>{t('SettingsDialog.Settings')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='flex min-h-0 flex-1 flex-col overflow-hidden'>
        {loading ? (
          <div className='grid min-h-0 flex-1 place-items-center'>
            <Spinner size='lg' />
          </div>
        ) : isMobile ? (
          <div className='flex min-h-0 flex-1 flex-col overflow-hidden'>
            <div className='sticky top-0 z-10 shrink-0 border-b border-border/60 bg-surface pb-3'>
              <Label className='sr-only'>{t('SettingsDialog.SectionPicker')}</Label>
              <Select
                selectedKey={tab}
                onSelectionChange={key => {
                  if (key != null) setTab(String(key) as SettingsTab)
                }}
                className='w-full'
                aria-label={t('SettingsDialog.SectionPicker')}
              >
                <Select.Trigger className='min-h-11 w-full'>
                  <Select.Value />
                  <Select.Indicator />
                </Select.Trigger>
                <Select.Popover>
                  <ListBox>
                    {visibleTabs.map(item => (
                      <ListBox.Item key={item.id} id={item.id} textValue={item.shortLabel}>
                        <span className='flex items-center gap-2.5'>
                          <span className='shrink-0 text-muted'>{item.icon}</span>
                          <span>{item.shortLabel}</span>
                        </span>
                      </ListBox.Item>
                    ))}
                  </ListBox>
                </Select.Popover>
              </Select>
            </div>
            <div className={panelClassName}>
              <PanelFade key={tab}>{renderPanel(tab)}</PanelFade>
            </div>
          </div>
        ) : (
          <Tabs.Root
            orientation='vertical'
            selectedKey={tab}
            onSelectionChange={key => setTab(String(key) as SettingsTab)}
            className={tabsRootClassName}
          >
            <Tabs.List aria-label={t('SettingsDialog.Settings')} className={tabsListClassName}>
              {visibleTabs.map(item => (
                <Tabs.Tab
                  key={item.id}
                  id={item.id}
                  className='min-h-9 items-start justify-start gap-2.5 px-3 py-1.5 text-left'
                >
                  <span className='mt-0.5'>{item.icon}</span>
                  <span className='text-wrap leading-snug'>{item.label}</span>
                </Tabs.Tab>
              ))}
            </Tabs.List>

            {visibleTabs.map(item => (
              <Tabs.Panel key={item.id} id={item.id} className={panelClassName}>
                <PanelFade>{renderPanel(item.id)}</PanelFade>
              </Tabs.Panel>
            ))}
          </Tabs.Root>
        )}
      </Modal.Body>
      <Modal.Footer className='shrink-0'>
        {isMobile ? (
          <div className='flex w-full flex-col gap-2'>
            <Button
              onPress={() => void handleResetDefaults()}
              isDisabled={loading || saving}
              variant='outline'
              className={`${footerButtonClassName || ''} w-full`.trim()}
            >
              {t('SettingsDialog.ResetToDefault')}
            </Button>
            <Button
              variant='primary'
              onPress={() => void handleSave()}
              isDisabled={loading || saving}
              className={`${footerButtonClassName || ''} w-full`.trim()}
            >
              {saving ? <Spinner size='sm' color='current' /> : t('Save')}
            </Button>
          </div>
        ) : (
          <>
            <Button
              onPress={() => void handleResetDefaults()}
              isDisabled={loading || saving}
              variant='outline'
              className={footerButtonClassName}
            >
              {t('SettingsDialog.ResetToDefault')}
            </Button>
            <Button
              onPress={onClose}
              isDisabled={saving}
              variant='secondary'
              className={footerButtonClassName}
              autoFocus
            >
              {t('Cancel')}
            </Button>
            <Button
              variant='primary'
              onPress={() => void handleSave()}
              isDisabled={loading || saving}
              className={footerButtonClassName}
            >
              {saving ? <Spinner size='sm' color='current' /> : t('Save')}
            </Button>
          </>
        )}
      </Modal.Footer>
    </AppDialog>
  )
}
