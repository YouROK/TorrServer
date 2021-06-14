import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import 'i18n'

import './index.css'
import App from './components/App'

ReactDOM.render(
  <StrictMode>
    <App />
  </StrictMode>,
  document.getElementById('root'),
)
