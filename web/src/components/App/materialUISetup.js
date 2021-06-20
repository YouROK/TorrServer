import useMediaQuery from '@material-ui/core/useMediaQuery'
import { createMuiTheme } from '@material-ui/core'
import { useMemo } from 'react'

// https://material-ui.com/ru/customization/default-theme/
export const darkTheme = createMuiTheme({
  palette: {
    type: 'dark',
    background: { paper: '#575757' },
  },
})
export const lightTheme = createMuiTheme({
  palette: {
    type: 'light',
    background: { paper: '#f1f1f1' },
  },
})

export const useMaterialUITheme = () => {
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')

  const materialUITheme = useMemo(
    () =>
      createMuiTheme({
        palette: {
          type: prefersDarkMode ? 'dark' : 'light',
          primary: { main: '#00a572' },
        },
        typography: { fontFamily: 'Open Sans, sans-serif' },
      }),
    [prefersDarkMode],
  )

  return materialUITheme
}
