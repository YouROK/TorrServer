import CssBaseline from '@material-ui/core/CssBaseline'
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core'
import { useEffect, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import { Menu as MenuIcon, Close as CloseIcon } from '@material-ui/icons'
import { getTorrServerHost } from 'utils/Hosts'
import TorrentList from 'components/TorrentList'
import DonateSnackbar from 'components/Donate'
import DonateDialog from 'components/Donate/DonateDialog'
import Div100vh from 'react-div-100vh'

import { AppWrapper, AppHeader } from './style'
import Sidebar from './Sidebar'

const baseTheme = createMuiTheme({
  overrides: { MuiCssBaseline: { '@global': { html: { WebkitFontSmoothing: 'auto' } } } },
  palette: { primary: { main: '#3fb57a' }, secondary: { main: '#FFA724' }, tonalOffset: 0.2 },
})

export default function App() {
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [isDonationDialogOpen, setIsDonationDialogOpen] = useState(false)
  const [tsVersion, setTSVersion] = useState('')

  useEffect(() => {
    fetch(`${getTorrServerHost()}/echo`)
      .then(resp => resp.text())
      .then(txt => {
        if (!txt.startsWith('<!DOCTYPE html>')) setTSVersion(txt)
      })
  }, [isDrawerOpen])

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
              TorrServer {tsVersion}
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
