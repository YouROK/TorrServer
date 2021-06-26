import { createMuiTheme, useMediaQuery } from '@material-ui/core'
import { useEffect, useMemo, useState } from 'react'

import { mainColors } from './colors'

const typography = { fontFamily: 'Open Sans, sans-serif' }

export const darkTheme = createMuiTheme({
  typography,
  palette: {
    type: 'dark',
    primary: { main: mainColors.dark.primary },
  },
})
export const lightTheme = createMuiTheme({
  typography,
  palette: {
    type: 'light',
    primary: { main: mainColors.light.primary },
  },
})

export const THEME_MODES = { LIGHT: 'light', DARK: 'dark', AUTO: 'auto' }

export const useMaterialUITheme = () => {
  const currentModeState = useMediaQuery('(prefers-color-scheme: dark)')
  const [isDarkMode, setIsDarkMode] = useState(currentModeState)
  const [currentThemeMode, setCurrentThemeMode] = useState(THEME_MODES.LIGHT)

  useEffect(() => {
    currentThemeMode === THEME_MODES.LIGHT && setIsDarkMode(false)
    currentThemeMode === THEME_MODES.DARK && setIsDarkMode(true)
    currentThemeMode === THEME_MODES.AUTO && setIsDarkMode(currentModeState)
  }, [currentModeState, currentThemeMode])

  const theme = isDarkMode ? THEME_MODES.DARK : THEME_MODES.LIGHT

  const muiTheme = useMemo(
    () =>
      createMuiTheme({
        typography,
        palette: {
          type: theme,
          primary: { main: mainColors[theme].primary },
        },
      }),
    [theme],
  )

  return [isDarkMode, currentThemeMode, setCurrentThemeMode, muiTheme]
}
