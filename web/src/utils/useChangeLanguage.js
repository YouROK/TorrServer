import { useTranslation } from 'react-i18next'

export default () => {
  const { i18n } = useTranslation()
  const currentLanguage =
    i18n.language === 'en-US' || i18n.language === 'en'
      ? 'en'
      : i18n.language === 'ru-RU' || i18n.language === 'ru'
      ? 'ru'
      : i18n.language

  return [currentLanguage, lang => i18n.changeLanguage(lang)]
}
