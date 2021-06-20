import useMediaQuery from '@material-ui/core/useMediaQuery'
import { createMuiTheme } from '@material-ui/core'
import { useMemo } from 'react'

import { mainColors } from './colors'

const primary = { main: mainColors.primary }
const typography = { fontFamily: 'Open Sans, sans-serif' }

// https://material-ui.com/ru/customization/default-theme/
export const darkTheme = createMuiTheme({
  typography,
  palette: {
    type: 'dark',
    background: { paper: '#575757' },
    primary,
  },
})
export const lightTheme = createMuiTheme({
  typography,
  palette: {
    type: 'light',
    background: { paper: '#f1f1f1' },
    primary,
  },
})

export const useMaterialUITheme = () => {
  const isDarkMode = useMediaQuery('(prefers-color-scheme: dark)')

  const muiTheme = useMemo(
    () =>
      createMuiTheme({
        typography,
        palette: {
          type: isDarkMode ? 'dark' : 'light',
          primary,
        },
      }),
    [isDarkMode],
  )

  return [isDarkMode, muiTheme]
}
