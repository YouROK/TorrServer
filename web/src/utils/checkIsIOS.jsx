import { detectApplePlatform } from './Utils'

export default () => {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') return false
  const { isIOS } = detectApplePlatform()
  return isIOS
}
