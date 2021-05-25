import CssBaseline from '@material-ui/core/CssBaseline'
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core'

import Appbar from './components/Appbar/index'

const baseTheme = createMuiTheme({
  overrides: {
    MuiCssBaseline: {
      '@global': {
        html: {
          WebkitFontSmoothing: 'auto',
        },
      },
    },
  },
  palette: {
    primary: {
      main: '#3fb57a',
    },
    secondary: {
      main: '#FFA724',
    },
    tonalOffset: 0.2,
  },
})

export default function App() {
  return (
    <MuiThemeProvider theme={baseTheme}>
      <CssBaseline />
      <Appbar />
    </MuiThemeProvider>
  )
}
