import { lazy, type ReactNode, Suspense, useEffect, useMemo, useState } from 'react'
import axios from 'axios'
import { Button, Dropdown, Spinner, Tooltip, useMediaQuery } from '@heroui/react'
import { Check, ChevronLeft, Menu, Moon, Palette, SortAsc, SortDesc, Sun, SunMoon, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { echoHost } from 'shared/api/hosts'
import { SUPPORTED_LANGS } from 'shared/i18n'
import useChangeLanguage from 'shared/lib/useChangeLanguage'
import useLaunchHandler from 'shared/lib/useLaunchHandler'
import { detectApplePlatform, isStandaloneApp } from 'shared/lib/platform'
import { useLocalJsonPref } from 'shared/hooks/useLocalPref'
import { useTorrentsQuery } from 'shared/hooks/useTorrentsQuery'
import { OPEN_SETTINGS_EVENT, type SettingsDeepLinkTab } from 'shared/lib/settingsEvents'
import { getStoredCredentials, logoutBasicAuth } from 'shared/api/authCredentials'
import { MEDIA_SHORT_VIEWPORT, queryMax } from 'shared/theme/breakpoints'
import { THEME_PALETTE_SWATCHES } from 'shared/theme/paletteSwatches'
import { THEME_MODES, THEME_PALETTE_IDS, useThemePreference, type ThemePalette } from 'shared/theme/useThemePreference'
import { TORRENT_CATEGORIES } from 'shared/torrent/categories'
import { TorrentsPage } from 'features/torrents'
import { iconBtn } from 'shared/ui/controlClasses'
import { iconMenu, iconNav, iconNavMobile } from 'shared/ui/iconProps'
import DialogErrorBoundary from 'shared/ui/DialogErrorBoundary'

import BottomNav from './BottomNav'
import Sidebar from './Sidebar'

const CommandPalette = lazy(() => import('./CommandPalette'))
const AddDialog = lazy(() => import('features/add/AddDialog'))
const MultiAddDialog = lazy(() => import('features/add/MultiAddDialog'))
const SearchDialog = lazy(() => import('features/search/SearchDialog'))
const SettingsDialog = lazy(() => import('features/settings/SettingsDialog'))
const AboutDialog = lazy(() => import('features/about/AboutDialog'))
const DonateDialog = lazy(() => import('features/donate/DonateDialog'))
const DonateSnackbar = lazy(() => import('features/donate/DonateSnackbar'))
const CloseServerDialog = lazy(() => import('features/system/CloseServerDialog'))
const RemoveAllDialog = lazy(() => import('features/system/RemoveAllDialog'))
const ServerStatusDialog = lazy(() => import('features/system/ServerStatusDialog'))
const CategoriesDrawer = lazy(() => import('features/categories/CategoriesDrawer'))
const PWAInstallationGuide = lazy(() => import('features/pwa/PWAInstallationGuide'))
const AndroidInstallBanner = lazy(() => import('features/pwa/AndroidInstallBanner'))

const lazyDialogFallback = (
  <div className='grid place-items-center p-4'>
    <Spinner size='sm' />
  </div>
)

const LANG_CYCLE = SUPPORTED_LANGS

/** Endonyms stay in their own language — no i18n lookup needed. */
const LANG_OPTIONS: { id: (typeof LANG_CYCLE)[number]; code: string; name: string }[] = [
  { id: 'en', code: 'EN', name: 'English' },
  { id: 'ru', code: 'RU', name: 'Русский' },
  { id: 'ua', code: 'UA', name: 'Українська' },
  { id: 'zh', code: 'ZH', name: '中文' },
  { id: 'bg', code: 'BG', name: 'Български' },
  { id: 'fr', code: 'FR', name: 'Français' },
  { id: 'ro', code: 'RO', name: 'Română' },
]

const SIDEBAR_OPEN_PX = 260
const SIDEBAR_COLLAPSED_PX = 60
const HEADER_HEIGHT = 'calc(60px + env(safe-area-inset-top, 0px))'
const HEADER_HEIGHT_SHORT = 'calc(44px + env(safe-area-inset-top, 0px))'

/**
 * App chrome: sidebar / bottom nav, library host, and lazy-loaded dialogs.
 * Owns offline echo probe, theme/language toggles, and settings deep-link events.
 */
export default function Shell() {
  const { t } = useTranslation()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isShortViewport = useMediaQuery(MEDIA_SHORT_VIEWPORT)

  const { preference: currentThemeMode, setPreference: updateThemeMode, palette, setPalette } = useThemePreference()
  const [currentLang, changeLang] = useChangeLanguage()
  const { launchSource, setLaunchSource, launchFiles, setLaunchFiles } = useLaunchHandler()
  const [sidebarOpen, setSidebarOpen] = useLocalJsonPref('sidebarOpen', true)

  const [torrServerVersion, setTorrServerVersion] = useState('')
  const [sortABC, setSortABC] = useLocalJsonPref('sortABC', false)
  const [globalCategoryFilter, setGlobalCategoryFilter] = useLocalJsonPref('categoryFilter', 'all')

  const [addOpen, setAddOpen] = useState(false)
  const [searchOpen, setSearchOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [settingsInitialTab, setSettingsInitialTab] = useState<SettingsDeepLinkTab | undefined>()
  const [aboutOpen, setAboutOpen] = useState(false)
  const [donateOpen, setDonateOpen] = useState(false)
  const [categoriesOpen, setCategoriesOpen] = useState(false)
  const [closeServerOpen, setCloseServerOpen] = useState(false)
  const [removeAllOpen, setRemoveAllOpen] = useState(false)
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false)
  const [serverStatusOpen, setServerStatusOpen] = useState(false)

  const { isLoading, isError } = useTorrentsQuery()
  const isOffline = isError
  const sidebarWidth = sidebarOpen ? SIDEBAR_OPEN_PX : SIDEBAR_COLLAPSED_PX

  useEffect(() => {
    axios
      .get(echoHost())
      .then(({ data }) => setTorrServerVersion(String(data)))
      .catch(() => setTorrServerVersion(''))
  }, [])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- open Add when OS share/launch hands off a torrent
    if (launchSource || launchFiles) setAddOpen(true)
  }, [launchSource, launchFiles])

  useEffect(() => {
    const openWithTab = (event: Event) => {
      setSettingsInitialTab((event as CustomEvent<SettingsDeepLinkTab>).detail)
      setSettingsOpen(true)
    }
    window.addEventListener(OPEN_SETTINGS_EVENT, openWithTab)
    return () => window.removeEventListener(OPEN_SETTINGS_EVENT, openWithTab)
  }, [])

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (!(event.metaKey || event.ctrlKey) || event.key.toLowerCase() !== 'k') return
      const target = event.target as HTMLElement | null
      if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) {
        return
      }
      event.preventDefault()
      setCommandPaletteOpen(true)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  const closeAdd = () => {
    setAddOpen(false)
    setLaunchSource(null)
  }

  const cycleTheme = () => {
    if (currentThemeMode === THEME_MODES.LIGHT) updateThemeMode(THEME_MODES.DARK)
    else if (currentThemeMode === THEME_MODES.DARK) updateThemeMode(THEME_MODES.AUTO)
    else updateThemeMode(THEME_MODES.LIGHT)
  }

  const isCategoryFilterActive = globalCategoryFilter !== 'all'
  const categoryFilterLabel =
    globalCategoryFilter === 'all'
      ? null
      : globalCategoryFilter === ''
        ? t('Uncategorized')
        : t(TORRENT_CATEGORIES.find(category => category.key === globalCategoryFilter)?.name ?? globalCategoryFilter)

  const navProps = {
    isOffline,
    isLoading,
    isCategoryFilterActive,
    onAdd: () => setAddOpen(true),
    onSearch: () => setSearchOpen(true),
    onCategories: () => setCategoriesOpen(true),
    onSettings: () => setSettingsOpen(true),
    onAbout: () => setAboutOpen(true),
    onDonate: () => setDonateOpen(true),
    onServerStatus: () => setServerStatusOpen(true),
    onCloseServer: () => setCloseServerOpen(true),
    onRemoveAll: () => setRemoveAllOpen(true),
    ...(getStoredCredentials() ? { onLogout: () => logoutBasicAuth() } : {}),
  }

  const ThemeIcon =
    currentThemeMode === THEME_MODES.LIGHT ? Sun : currentThemeMode === THEME_MODES.DARK ? Moon : SunMoon
  const themeModeLabel =
    currentThemeMode === THEME_MODES.LIGHT
      ? t('ThemeLight')
      : currentThemeMode === THEME_MODES.DARK
        ? t('ThemeDark')
        : t('ThemeAuto')
  const paletteLabels = useMemo(
    () =>
      Object.fromEntries(
        THEME_PALETTE_IDS.map(id => [id, t(`ThemePalette${id.charAt(0).toUpperCase()}${id.slice(1)}`)]),
      ) as Record<ThemePalette, string>,
    [t],
  )
  const SortIcon = sortABC ? SortAsc : SortDesc
  const headerHeight = isShortViewport ? HEADER_HEIGHT_SHORT : HEADER_HEIGHT

  return (
    <div
      className='grid min-h-0 w-full overflow-hidden bg-background'
      style={{
        height: '100%',
        minHeight: 0,
        gridTemplateRows: isMobile ? `${headerHeight} minmax(0, 1fr) auto` : `${headerHeight} minmax(0, 1fr)`,
        gridTemplateColumns: isMobile ? '1fr' : `${sidebarWidth}px 1fr`,
        gridTemplateAreas: isMobile ? '"header" "content" "nav"' : '"header header" "sidebar content"',
        transition: 'grid-template-columns 200ms ease',
      }}
    >
      <header
        className={`flex items-center bg-app-header text-app-header-foreground pl-[env(safe-area-inset-left,0px)] pr-[env(safe-area-inset-right,0px)] ${
          isShortViewport
            ? 'gap-1 px-1 pt-[env(safe-area-inset-top,0px)]'
            : 'gap-2 px-2 pt-[env(safe-area-inset-top,0px)]'
        }`}
        style={{ gridArea: 'header', minHeight: headerHeight }}
      >
        {!isMobile ? (
          <HeaderIconButton
            label={sidebarOpen ? t('CollapseSidebar') : t('ExpandSidebar')}
            onPress={() => setSidebarOpen(!sidebarOpen)}
          >
            {sidebarOpen ? <ChevronLeft {...iconNavMobile} /> : <Menu {...iconNavMobile} />}
          </HeaderIconButton>
        ) : null}

        <h1
          className={`flex min-w-0 flex-1 items-center gap-2 truncate font-semibold ${isShortViewport ? 'text-base' : 'text-lg'}`}
          title={torrServerVersion ? `TorrServer ${torrServerVersion}` : 'TorrServer'}
        >
          <span className='truncate'>
            TorrServer
            {!isMobile && torrServerVersion ? (
              <span className='font-normal text-app-header-foreground/70'>
                {' '}
                {torrServerVersion.includes('-') ? torrServerVersion.split('-')[0] : torrServerVersion}
              </span>
            ) : null}
          </span>
          {categoryFilterLabel ? (
            <button
              type='button'
              onClick={() => setGlobalCategoryFilter('all')}
              className='inline-flex min-h-11 shrink-0 items-center gap-1 rounded-full bg-white/15 px-2.5 py-1 text-xs font-medium hover-fine:bg-white/25'
              aria-label={t('ClearCategoryFilter')}
            >
              <span className='max-w-[9rem] truncate'>{categoryFilterLabel}</span>
              <X size={13} strokeWidth={1.75} aria-hidden />
            </button>
          ) : null}
        </h1>

        <HeaderIconButton label={sortABC ? t('SortByDate') : t('SortByName')} onPress={() => setSortABC(!sortABC)}>
          <SortIcon {...iconNav} />
        </HeaderIconButton>

        <HeaderIconButton label={`${t('Theme')}: ${themeModeLabel}`} onPress={cycleTheme}>
          <ThemeIcon {...iconNav} />
        </HeaderIconButton>

        <PaletteMenu current={palette} label={t('ThemePalette')} onChange={setPalette} labels={paletteLabels} />

        <LanguageMenu
          currentLang={LANG_CYCLE.includes(currentLang as (typeof LANG_CYCLE)[number]) ? currentLang : 'en'}
          label={t('Language')}
          onChange={changeLang}
        />
      </header>

      {!isMobile ? (
        <aside
          className='flex h-full min-h-0 flex-col overflow-hidden'
          style={{ gridArea: 'sidebar', width: sidebarWidth, transition: 'width 200ms ease' }}
        >
          <Sidebar {...navProps} collapsed={!sidebarOpen} />
        </aside>
      ) : null}

      <main
        className='min-h-0 min-w-0 overflow-auto overscroll-y-contain bg-background [-webkit-overflow-scrolling:touch]'
        style={{ gridArea: 'content' }}
      >
        <TorrentsPage
          sortABC={sortABC}
          sortCategory={globalCategoryFilter}
          onAdd={() => setAddOpen(true)}
          onClearCategory={() => setGlobalCategoryFilter('all')}
        />
      </main>

      {isMobile ? (
        <div className='min-h-0 shrink-0' style={{ gridArea: 'nav' }}>
          <BottomNav {...navProps} />
        </div>
      ) : null}

      <Suspense fallback={lazyDialogFallback}>
        <DialogErrorBoundary onClose={closeAdd}>
          <AddDialog open={addOpen && !launchFiles} onClose={closeAdd} initialSource={launchSource} />
        </DialogErrorBoundary>
        {launchFiles ? (
          <DialogErrorBoundary
            onClose={() => {
              setLaunchFiles(null)
              setAddOpen(false)
            }}
          >
            <MultiAddDialog
              files={launchFiles}
              open
              onClose={() => {
                setLaunchFiles(null)
                setAddOpen(false)
              }}
            />
          </DialogErrorBoundary>
        ) : null}
        <DialogErrorBoundary onClose={() => setSearchOpen(false)}>
          <SearchDialog open={searchOpen} onClose={() => setSearchOpen(false)} />
        </DialogErrorBoundary>
        <DialogErrorBoundary
          onClose={() => {
            setSettingsOpen(false)
            setSettingsInitialTab(undefined)
          }}
        >
          <SettingsDialog
            open={settingsOpen}
            onClose={() => {
              setSettingsOpen(false)
              setSettingsInitialTab(undefined)
            }}
            initialTab={settingsInitialTab}
          />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setAboutOpen(false)}>
          <AboutDialog
            open={aboutOpen}
            onClose={() => setAboutOpen(false)}
            onOpenServerStatus={() => setServerStatusOpen(true)}
            onOpenDonate={() => setDonateOpen(true)}
          />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setDonateOpen(false)}>
          <DonateDialog open={donateOpen} onClose={() => setDonateOpen(false)} />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setCloseServerOpen(false)}>
          <CloseServerDialog open={closeServerOpen} onClose={() => setCloseServerOpen(false)} />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setRemoveAllOpen(false)}>
          <RemoveAllDialog open={removeAllOpen} onClose={() => setRemoveAllOpen(false)} />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setServerStatusOpen(false)}>
          <ServerStatusDialog open={serverStatusOpen} onClose={() => setServerStatusOpen(false)} />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setCommandPaletteOpen(false)}>
          {commandPaletteOpen ? (
            <CommandPalette
              key='cmdk'
              open
              onClose={() => setCommandPaletteOpen(false)}
              onAdd={() => setAddOpen(true)}
              onSearch={() => setSearchOpen(true)}
              onAbout={() => setAboutOpen(true)}
              onDonate={() => setDonateOpen(true)}
              onServerStatus={() => setServerStatusOpen(true)}
              onToggleTheme={cycleTheme}
            />
          ) : null}
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setCategoriesOpen(false)}>
          <CategoriesDrawer
            open={categoriesOpen}
            onClose={() => setCategoriesOpen(false)}
            selectedCategory={globalCategoryFilter}
            onSelectCategory={setGlobalCategoryFilter}
          />
        </DialogErrorBoundary>
        {detectApplePlatform().isIOS && !isStandaloneApp ? <PWAInstallationGuide /> : null}
        {!detectApplePlatform().isIOS && !isStandaloneApp ? <AndroidInstallBanner /> : null}
        {/* Hide donate while iOS install guide is showing — both compete for the bottom band. */}
        {!(detectApplePlatform().isIOS && !isStandaloneApp) ? (
          <DonateSnackbar onSupport={() => setDonateOpen(true)} />
        ) : null}
      </Suspense>
    </div>
  )
}

function PaletteMenu({
  current,
  label,
  onChange,
  labels,
}: {
  current: ThemePalette
  label: string
  onChange: (palette: ThemePalette) => void
  labels: Record<ThemePalette, string>
}) {
  return (
    <Dropdown>
      <Dropdown.Trigger>
        <Button
          variant='ghost'
          isIconOnly
          className={`${iconBtn} text-app-header-foreground hover-fine:bg-white/10`}
          aria-label={`${label}: ${labels[current]}`}
        >
          <span className='inline-flex size-full items-center justify-center [&>svg]:m-0 [&>svg]:block'>
            <Palette {...iconNav} aria-hidden />
          </span>
        </Button>
      </Dropdown.Trigger>
      <Dropdown.Popover placement='bottom end' className='max-h-[min(70dvh,28rem)] min-w-[13rem] overflow-y-auto'>
        <Dropdown.Menu aria-label={label}>
          {THEME_PALETTE_IDS.map(id => {
            const selected = id === current
            const swatch = THEME_PALETTE_SWATCHES[id]
            return (
              <Dropdown.Item key={id} textValue={labels[id]} onPress={() => onChange(id)} className='min-h-11 gap-2'>
                <span
                  className='size-4 shrink-0 rounded-sm border border-black/10'
                  aria-hidden
                  style={{
                    background: `linear-gradient(135deg, ${swatch.header} 0 52%, ${swatch.accent} 52% 100%)`,
                  }}
                />
                <span className='min-w-0 flex-1 truncate text-sm'>{labels[id]}</span>
                {selected ? (
                  <Check {...iconMenu} className='shrink-0 text-accent' aria-hidden />
                ) : (
                  <span className='size-4 shrink-0' aria-hidden />
                )}
              </Dropdown.Item>
            )
          })}
        </Dropdown.Menu>
      </Dropdown.Popover>
    </Dropdown>
  )
}

function LanguageMenu({
  currentLang,
  label,
  onChange,
}: {
  currentLang: string
  label: string
  onChange: (lang: string) => void
}) {
  const current = LANG_OPTIONS.find(option => option.id === currentLang) ?? LANG_OPTIONS[0]

  return (
    <Dropdown>
      <Dropdown.Trigger>
        <Button
          variant='ghost'
          className={`${iconBtn} text-xs font-semibold tracking-wide text-app-header-foreground hover-fine:bg-white/10`}
          aria-label={label}
        >
          {current.code}
        </Button>
      </Dropdown.Trigger>
      <Dropdown.Popover placement='bottom end' className='min-w-[12rem]'>
        <Dropdown.Menu aria-label={label}>
          {LANG_OPTIONS.map(option => {
            const selected = option.id === current.id
            return (
              <Dropdown.Item
                key={option.id}
                textValue={`${option.code} ${option.name}`}
                onPress={() => onChange(option.id)}
                className='min-h-11 gap-2'
              >
                <span className='w-7 shrink-0 text-xs font-semibold tabular-nums'>{option.code}</span>
                <span className='min-w-0 flex-1 truncate text-sm'>{option.name}</span>
                {selected ? (
                  <Check {...iconMenu} className='shrink-0 text-accent' aria-hidden />
                ) : (
                  <span className='size-4 shrink-0' aria-hidden />
                )}
              </Dropdown.Item>
            )
          })}
        </Dropdown.Menu>
      </Dropdown.Popover>
    </Dropdown>
  )
}

function HeaderIconButton({ label, onPress, children }: { label: string; onPress: () => void; children: ReactNode }) {
  return (
    <Tooltip>
      <Tooltip.Trigger>
        <Button
          variant='ghost'
          isIconOnly
          className={`${iconBtn} text-app-header-foreground hover-fine:bg-white/10`}
          aria-label={label}
          onPress={onPress}
        >
          <span className='inline-flex size-full items-center justify-center [&>svg]:m-0 [&>svg]:block'>
            {children}
          </span>
        </Button>
      </Tooltip.Trigger>
      <Tooltip.Content>{label}</Tooltip.Content>
    </Tooltip>
  )
}
