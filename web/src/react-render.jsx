import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import { QueryClientProvider, QueryClient } from 'react-query'

import App from './components/App'
import './i18n'

export const applicationMount = (id = 'root') => {
  const queryClient = new QueryClient()

  ReactDOM.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </StrictMode>,
    typeof id === 'string' ? document.getElementById(id): id,
  )
}
