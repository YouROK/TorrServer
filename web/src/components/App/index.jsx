import useMediaQuery from '@material-ui/core/useMediaQuery'
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core'
import CssBaseline from '@material-ui/core/CssBaseline'
import { useEffect, useMemo, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import { Menu as MenuIcon, Close as CloseIcon } from '@material-ui/icons'
import { echoHost } from 'utils/Hosts'
import Div100vh from 'react-div-100vh'
import axios from 'axios'
import TorrentList from 'components/TorrentList'
import DonateSnackbar from 'components/Donate'
import DonateDialog from 'components/Donate/DonateDialog'

import { AppWrapper, AppHeader } from './style'
import Sidebar from './Sidebar'

export default function App() {
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [isDonationDialogOpen, setIsDonationDialogOpen] = useState(false)
  const [torrServerVersion, setTorrServerVersion] = useState('')
  // https://material-ui.com/ru/customization/palette/
  const baseTheme = useMemo(
    () =>
      createMuiTheme({
        overrides: { MuiCssBaseline: { '@global': { html: { WebkitFontSmoothing: 'auto' } } } },
        palette: {
          type: prefersDarkMode ? 'dark' : 'light',
          primary: { main: '#00a572' },
          secondary: { main: '#ffa724' },
          tonalOffset: 0.2,
        },
      }),
    [prefersDarkMode],
  )

  useEffect(() => {
    axios.get(echoHost()).then(({ data }) => setTorrServerVersion(data))
  }, [])

  return (
    <MuiThemeProvider theme={baseTheme}>
      <CssBaseline />

      {/* Div100vh - iOS WebKit fix  */}
      <Div100vh>
        <AppWrapper>
          <AppHeader>
            <IconButton
              style={{ marginRight: '20px' }}
              color='inherit'
              onClick={() => setIsDrawerOpen(!isDrawerOpen)}
              edge='start'
            >
              {isDrawerOpen ? <CloseIcon /> : <MenuIcon />}
            </IconButton>

            <Typography variant='h6' noWrap>
              TorrServer {torrServerVersion}
            </Typography>
          </AppHeader>

          <Sidebar isDrawerOpen={isDrawerOpen} setIsDonationDialogOpen={setIsDonationDialogOpen} />

          <TorrentList />

          {isDonationDialogOpen && <DonateDialog onClose={() => setIsDonationDialogOpen(false)} />}
          {!JSON.parse(localStorage.getItem('snackbarIsClosed')) && <DonateSnackbar />}
        </AppWrapper>
      </Div100vh>
    </MuiThemeProvider>
  )
}
