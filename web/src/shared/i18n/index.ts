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

export const SUPPORTED_LANGS = ['en', 'ru', 'ua', 'zh', 'bg', 'fr', 'ro'] as const

/**
 * i18n bootstrap: `load: 'languageOnly'` + `nonExplicitSupportedLngs` so `en-US`
 * resolves to `en`. Detection prefers localStorage over navigator.
 */
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    supportedLngs: [...SUPPORTED_LANGS],
    nonExplicitSupportedLngs: true,
    load: 'languageOnly',
    fallbackLng: 'en',
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
    },
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
