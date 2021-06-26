import { createMuiTheme, useMediaQuery } from '@material-ui/core'
import { useEffect, useMemo, useState } from 'react'

import { mainColors } from './colors'

export const THEME_MODES = { LIGHT: 'light', DARK: 'dark', AUTO: 'auto' }

const typography = { fontFamily: 'Open Sans, sans-serif' }

export const darkTheme = createMuiTheme({
  typography,
  palette: {
    type: THEME_MODES.DARK,
    primary: { main: mainColors.dark.primary },
  },
})
export const lightTheme = createMuiTheme({
  typography,
  palette: {
    type: THEME_MODES.LIGHT,
    primary: { main: mainColors.light.primary },
  },
})

export const useMaterialUITheme = () => {
  const savedThemeMode = localStorage.getItem('themeMode')
  const isSystemModeDark = useMediaQuery('(prefers-color-scheme: dark)')
  const [isDarkMode, setIsDarkMode] = useState(savedThemeMode === 'dark' || isSystemModeDark)
  const [currentThemeMode, setCurrentThemeMode] = useState(savedThemeMode || THEME_MODES.LIGHT)

  useEffect(() => {
    currentThemeMode === THEME_MODES.LIGHT && setIsDarkMode(false)
    currentThemeMode === THEME_MODES.DARK && setIsDarkMode(true)
    currentThemeMode === THEME_MODES.AUTO && setIsDarkMode(isSystemModeDark)
  }, [isSystemModeDark, currentThemeMode])

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
