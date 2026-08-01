import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import { QueryClientProvider, QueryClient } from 'react-query'

import App from './components/App'
import { PlaybackProvider } from './playback/PlaybackContext'
import 'i18n'

const queryClient = new QueryClient()

ReactDOM.render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <PlaybackProvider>
        <App />
      </PlaybackProvider>
    </QueryClientProvider>
  </StrictMode>,
  document.getElementById('root'),
)
