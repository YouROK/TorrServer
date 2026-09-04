import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

function applyDocumentLang(lang: string) {
  document.documentElement.lang = lang.substring(0, 2)
}

export default function useChangeLanguage(): [string, (lang: string) => void] {
  const { i18n } = useTranslation()
  const currentLanguage = i18n.language.substring(0, 2)

  useEffect(() => {
    applyDocumentLang(i18n.language)
  }, [i18n.language])

  return [
    currentLanguage,
    lang => {
      applyDocumentLang(lang)
      void i18n.changeLanguage(lang)
    },
  ]
}
