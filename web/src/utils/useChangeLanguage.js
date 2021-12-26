import { useTranslation } from 'react-i18next'

export default () => {
  const { i18n } = useTranslation()
  const currentLanguage = i18n.language.substr(0, 2)

  return [currentLanguage, lang => i18n.changeLanguage(lang)]
}
