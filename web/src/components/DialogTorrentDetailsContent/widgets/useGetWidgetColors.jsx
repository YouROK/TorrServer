import { DarkModeContext } from 'components/App'
import { useContext } from 'react'
import { THEME_MODES } from 'style/materialUISetup'

const { LIGHT, DARK } = THEME_MODES

const colors = {
  light: {
    downloadSpeed: { iconBGColor: '#118f00', valueBGColor: '#13a300' },
    uploadSpeed: { iconBGColor: '#0146ad', valueBGColor: '#0058db' },
    peers: { iconBGColor: '#cdc118', valueBGColor: '#d8cb18' },
    piecesCount: { iconBGColor: '#b6c95e', valueBGColor: '#c0d076' },
    piecesLength: { iconBGColor: '#0982c8', valueBGColor: '#098cd7' },
    status: { iconBGColor: '#aea25b', valueBGColor: '#b4aa6e' },
    size: { iconBGColor: '#9b01ad', valueBGColor: '#ac03bf' },
    category: { iconBGColor: '#914820', valueBGColor: '#c9632c' },
  },
  dark: {
    downloadSpeed: { iconBGColor: '#0c6600', valueBGColor: '#0d7000' },
    uploadSpeed: { iconBGColor: '#003f9e', valueBGColor: '#0047b3' },
    peers: { iconBGColor: '#a69c11', valueBGColor: '#b4a913' },
    piecesCount: { iconBGColor: '#8da136', valueBGColor: '#99ae3d' },
    piecesLength: { iconBGColor: '#07659c', valueBGColor: '#0872af' },
    status: { iconBGColor: '#938948', valueBGColor: '#9f9450' },
    size: { iconBGColor: '#81008f', valueBGColor: '#9102a1' },
    category: { iconBGColor: '#914820', valueBGColor: '#c9632c' },
  },
}

export default function useGetWidgetColors(widgetName) {
  const { isDarkMode } = useContext(DarkModeContext)
  const widgetColors = colors[isDarkMode ? DARK : LIGHT][widgetName]

  return widgetColors
}
