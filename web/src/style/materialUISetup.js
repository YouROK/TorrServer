import { createTheme, useMediaQuery } from '@material-ui/core'
import { useEffect, useMemo, useState } from 'react'

import { mainColors, themeColors } from './colors'

export const THEME_MODES = { LIGHT: 'light', DARK: 'dark', AUTO: 'auto' }

const typography = { fontFamily: 'Open Sans, sans-serif' }

export const darkTheme = createTheme({
  typography,
  palette: {
    type: THEME_MODES.DARK,
    primary: { main: mainColors.dark.primary },
    secondary: { main: mainColors.dark.secondary },
  },
})
export const lightTheme = createTheme({
  typography,
  palette: {
    type: THEME_MODES.LIGHT,
    primary: { main: mainColors.light.primary },
    secondary: { main: mainColors.light.secondary },
  },
})

export const useMaterialUITheme = () => {
  const savedThemeMode = localStorage.getItem('themeMode')
  const isSystemModeDark = useMediaQuery('(prefers-color-scheme: dark)')
  const [isDarkMode, setIsDarkMode] = useState(savedThemeMode === 'dark' || isSystemModeDark)
  const [currentThemeMode, setCurrentThemeMode] = useState(savedThemeMode || THEME_MODES.AUTO)

  const updateThemeMode = mode => {
    setCurrentThemeMode(mode)
    localStorage.setItem('themeMode', mode)
  }

  useEffect(() => {
    currentThemeMode === THEME_MODES.LIGHT && setIsDarkMode(false)
    currentThemeMode === THEME_MODES.DARK && setIsDarkMode(true)
    currentThemeMode === THEME_MODES.AUTO && setIsDarkMode(isSystemModeDark)
  }, [isSystemModeDark, currentThemeMode])

  const theme = isDarkMode ? THEME_MODES.DARK : THEME_MODES.LIGHT

  const muiTheme = useMemo(
    () =>
      createTheme({
        typography,
        palette: {
          type: theme,
          primary: { main: mainColors[theme].primary },
          secondary: { main: mainColors[theme].secondary },
        },
        overrides: {
          MuiTypography: {
            h6: {
              fontSize: '1.0rem',
            },
          },
          MuiPaper: {
            root: {
              backgroundColor: themeColors[theme].app.paperColor,
            },
          },
          MuiInputBase: {
            input: {
              color: mainColors[theme].labels,
            },
          },
          // https://material-ui.com/ru/api/form-control-label/
          MuiFormControlLabel: {
            labelPlacementStart: {
              display: 'flex',
              justifyContent: 'space-between',
              marginStart: 0,
              marginTop: 6,
              marginBottom: 2,
            },
          },
          MuiInputLabel: {
            root: {
              color: mainColors[theme].labels,
              marginBottom: 8,
              '&$focused': {
                color: mainColors[theme].labels,
              },
            },
          },
          MuiFormGroup: {
            root: {
              '& .MuiFormHelperText-root': {
                marginTop: -8,
              },
            },
          },
        },
      }),
    [theme],
  )

  return [isDarkMode, currentThemeMode, updateThemeMode, muiTheme]
}
