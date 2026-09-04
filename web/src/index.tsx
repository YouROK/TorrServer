import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import App from './app/App'
import ErrorBoundary from './app/ErrorBoundary'
import { installAppHeight } from 'shared/lib/appHeight'
import { bootstrapTelegramWebApp } from 'shared/lib/telegramWebApp'
import 'shared/i18n'
import './index.css'

// Before React paint — iOS PWA mis-resolves 100dvh; pin --app-height early (master used 100vh).
installAppHeight()
// Mini App SDK only when ?tg=1 — do not await (blocked telegram.org must not delay UI).
void bootstrapTelegramWebApp()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
})

const rootElement = document.getElementById('root')

if (rootElement) {
  createRoot(rootElement).render(
    <StrictMode>
      <ErrorBoundary>
        <QueryClientProvider client={queryClient}>
          <App />
        </QueryClientProvider>
      </ErrorBoundary>
    </StrictMode>,
  )
}
