import useMediaQuery from '@material-ui/core/useMediaQuery'
import { createMuiTheme } from '@material-ui/core'
import { useMemo } from 'react'

import { mainColors } from './colors'

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
  const isDarkMode = useMediaQuery('(prefers-color-scheme: dark)')

  const muiTheme = useMemo(
    () =>
      createMuiTheme({
        palette: {
          type: isDarkMode ? 'dark' : 'light',
          primary: { main: mainColors.primary },
        },
        typography: { fontFamily: 'Open Sans, sans-serif' },
      }),
    [isDarkMode],
  )

  return [isDarkMode, muiTheme]
}
