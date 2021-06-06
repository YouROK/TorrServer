import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import { I18nextProvider } from 'react-i18next'
import i18n from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

import './index.css'
import App from './App'

i18n.use(LanguageDetector).init({
  lng: 'en', // default
  fallbackLng: 'en', // use en if detected lng is not available
  keySeparator: false, // we do not use keys in form messages.welcome
  interpolation: {
    escapeValue: false, // react already safes from xss
  },
  resources: {
    en: {
      // eslint-disable-next-line global-require
      translations: require('./locales/en/translation.json'),
    },
    ru: {
      // eslint-disable-next-line global-require
      translations: require('./locales/ru/translation.json'),
    },
  },
  ns: ['translations'],
  defaultNS: 'translations',
})

i18n.languages = ['en', 'ru']

ReactDOM.render(
  <StrictMode>
    <I18nextProvider i18n={i18n}>
      <App />
    </I18nextProvider>
  </StrictMode>,
  document.getElementById('root'),
)
