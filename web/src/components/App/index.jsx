import CssBaseline from '@material-ui/core/CssBaseline'
import { createContext, useEffect, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import {
  Menu as MenuIcon,
  Close as CloseIcon,
  Brightness4 as Brightness4Icon,
  Brightness5 as Brightness5Icon,
  BrightnessAuto as BrightnessAutoIcon,
  Sort as SortIcon,
  SortByAlpha as SortByAlphaIcon,
} from '@material-ui/icons'
import { echoHost } from 'utils/Hosts'
import Div100vh from 'react-div-100vh'
import axios from 'axios'
import TorrentList from 'components/TorrentList'
import DonateSnackbar from 'components/Donate'
import DonateDialog from 'components/Donate/DonateDialog'
import useChangeLanguage from 'utils/useChangeLanguage'
import { ThemeProvider as MuiThemeProvider } from '@material-ui/core/styles'
import { ThemeProvider as StyledComponentsThemeProvider } from 'styled-components'
import { useQuery } from 'react-query'
import { getTorrents, isStandaloneApp } from 'utils/Utils'
import GlobalStyle from 'style/GlobalStyle'
import { /* lightTheme, */ THEME_MODES, useMaterialUITheme } from 'style/materialUISetup'
import getStyledComponentsTheme from 'style/getStyledComponentsTheme'
import checkIsIOS from 'utils/checkIsIOS'

import { AppWrapper, AppHeader, HeaderToggle, StyledIconButton } from './style'
import Sidebar from './Sidebar'
import PWAFooter from './PWAFooter'
import { PWAInstallationGuide } from './PWAInstallationGuide'

const snackbarIsClosed = JSON.parse(localStorage.getItem('snackbarIsClosed'))

export const DarkModeContext = createContext()

export default function App() {
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [isDonationDialogOpen, setIsDonationDialogOpen] = useState(false)
  const [torrServerVersion, setTorrServerVersion] = useState('')

  const [isDarkMode, currentThemeMode, updateThemeMode, muiTheme] = useMaterialUITheme()
  const [currentLang, changeLang] = useChangeLanguage()
  const [isOffline, setIsOffline] = useState(false)
  const [globalCategoryFilter, setGlobalFilterCategory] = useState('all')
  const { data: torrents, isLoading } = useQuery('torrents', getTorrents, {
    retry: 1,
    refetchInterval: 1000,
    onError: () => setIsOffline(true),
    onSuccess: () => setIsOffline(false),
  })
  const [sortABC, setSortABC] = useState(false)
  const handleClickSortABC = () => setSortABC(true)
  const handleClickSortDate = () => setSortABC(false)

  useEffect(() => {
    axios.get(echoHost()).then(({ data }) => setTorrServerVersion(data))
  }, [])

  return (
    <>
      <GlobalStyle />

      <DarkModeContext.Provider value={{ isDarkMode }}>
        <MuiThemeProvider theme={muiTheme}>
          <StyledComponentsThemeProvider
            theme={getStyledComponentsTheme(isDarkMode ? THEME_MODES.DARK : THEME_MODES.LIGHT)}
          >
            <CssBaseline />

            {/* Div100vh - iOS WebKit fix  */}
            <Div100vh>
              <AppWrapper>
                <AppHeader>
                  <StyledIconButton edge='start' color='inherit' onClick={() => setIsDrawerOpen(!isDrawerOpen)}>
                    {isDrawerOpen ? <CloseIcon /> : <MenuIcon />}
                  </StyledIconButton>

                  <Typography variant='h6' noWrap>
                    TorrServer {torrServerVersion}
                  </Typography>

                  <div
                    style={{ justifySelf: 'end', display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '10px' }}
                  >
                    <HeaderToggle onClick={() => (sortABC === true ? handleClickSortDate() : handleClickSortABC())}>
                      {sortABC === true ? <SortByAlphaIcon /> : <SortIcon />}
                    </HeaderToggle>

                    <HeaderToggle
                      onClick={() => {
                        if (currentThemeMode === THEME_MODES.LIGHT) updateThemeMode(THEME_MODES.DARK)
                        if (currentThemeMode === THEME_MODES.DARK) updateThemeMode(THEME_MODES.AUTO)
                        if (currentThemeMode === THEME_MODES.AUTO) updateThemeMode(THEME_MODES.LIGHT)
                      }}
                    >
                      {currentThemeMode === THEME_MODES.LIGHT ? (
                        <Brightness5Icon />
                      ) : currentThemeMode === THEME_MODES.DARK ? (
                        <Brightness4Icon />
                      ) : (
                        <BrightnessAutoIcon />
                      )}
                    </HeaderToggle>

                    <HeaderToggle
                      onClick={() =>
                        currentLang === 'en'
                          ? changeLang('ru')
                          : currentLang === 'ru'
                          ? changeLang('ua')
                          : currentLang === 'ua'
                          ? changeLang('zh')
                          : currentLang === 'zh'
                          ? changeLang('bg')
                          : currentLang === 'bg'
                          ? changeLang('fr')
                          : changeLang('en')
                      }
                    >
                      {currentLang.toUpperCase()}
                    </HeaderToggle>
                  </div>
                </AppHeader>

                <Sidebar
                  isOffline={isOffline}
                  isLoading={isLoading}
                  isDrawerOpen={isDrawerOpen}
                  setIsDonationDialogOpen={setIsDonationDialogOpen}
                  setGlobalFilterCategory={setGlobalFilterCategory}
                />

                <TorrentList
                  isOffline={isOffline}
                  torrents={torrents}
                  isLoading={isLoading}
                  sortABC={sortABC}
                  sortCategory={globalCategoryFilter}
                />

                <PWAFooter
                  isOffline={isOffline}
                  isLoading={isLoading}
                  setIsDonationDialogOpen={setIsDonationDialogOpen}
                />

                {/* <MuiThemeProvider theme={lightTheme}> */}
                {isDonationDialogOpen && <DonateDialog onClose={() => setIsDonationDialogOpen(false)} />}
                {/* </MuiThemeProvider> */}

                {snackbarIsClosed ? checkIsIOS() && !isStandaloneApp && <PWAInstallationGuide /> : <DonateSnackbar />}
              </AppWrapper>
            </Div100vh>
          </StyledComponentsThemeProvider>
        </MuiThemeProvider>
      </DarkModeContext.Provider>
    </>
  )
}
