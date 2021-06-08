import { useTranslation } from 'react-i18next'

export default () => {
  const { i18n } = useTranslation()
  return [i18n.language, lang => i18n.changeLanguage(lang)]
}
