import React from 'react'
import CssBaseline from '@material-ui/core/CssBaseline'
import Appbar from './components/Appbar.js'
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core'

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
        <React.Fragment>
            <MuiThemeProvider theme={baseTheme}>
                <CssBaseline />
                <Appbar />
            </MuiThemeProvider>
        </React.Fragment>
    )
}
