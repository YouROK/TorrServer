import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import { I18nextProvider } from 'react-i18next'
import i18n from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import XHR from 'i18next-xhr-backend'

import './index.css'
import App from './App'
import translationEng from './locales/en/translation.json'
import translationRus from './locales/ru/translation.json'

i18n
  .use(XHR)
  .use(LanguageDetector)
  .init({
    lng: 'ru', // default
    fallbackLng: 'en', // use en if detected lng is not available
    keySeparator: false, // we do not use keys in form messages.welcome
    interpolation: {
      escapeValue: false, // react already safes from xss
    },
    resources: {
      en: {
        translations: translationEng,
      },
      ru: {
        translations: translationRus,
      },
    },
    ns: ['translations'],
    defaultNS: 'translations',
  })

ReactDOM.render(
  <StrictMode>
    <I18nextProvider i18n={i18n}>
      <App />
    </I18nextProvider>
  </StrictMode>,
  document.getElementById('root'),
)
