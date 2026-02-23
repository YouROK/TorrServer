import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import translationEN from 'locales/en/translation.json'
import translationRU from 'locales/ru/translation.json'
import translationUA from 'locales/ua/translation.json'
import translationZH from 'locales/zh/translation.json'
import translationBG from 'locales/bg/translation.json'
import translationFR from 'locales/fr/translation.json'
import translationRO from 'locales/ro/translation.json'

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en', // default language will be used if none of declared lanuages detected (en, ru)
    interpolation: { escapeValue: false }, // react already safes from xss
    resources: {
      en: { translation: translationEN },
      ru: { translation: translationRU },
      ua: { translation: translationUA },
      zh: { translation: translationZH },
      bg: { translation: translationBG },
      fr: { translation: translationFR },
      ro: { translation: translationRO },
    },
  })

export default i18n
