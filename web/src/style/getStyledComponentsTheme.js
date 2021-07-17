import { mainColors, themeColors } from './colors'

export default type => ({ ...themeColors[type], ...mainColors[type] })
