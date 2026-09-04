import { createContext, useCallback, useContext, useMemo, type ReactNode } from 'react'
import { toast } from 'sonner'

export type ToastSeverity = 'info' | 'success' | 'warning' | 'error'

export interface ToastOptions {
  message: string
  severity?: ToastSeverity
  autoHideDuration?: number
}

interface AppSnackbarContextValue {
  showToast: (options: ToastOptions | string) => void
}

const AppSnackbarContext = createContext<AppSnackbarContextValue | null>(null)

export function useAppToast() {
  const ctx = useContext(AppSnackbarContext)
  if (!ctx) {
    throw new Error('useAppToast must be used within AppSnackbarProvider')
  }
  return ctx
}

export function useOptionalAppToast() {
  return useContext(AppSnackbarContext)
}

function showToastMessage(options: ToastOptions | string) {
  const next = typeof options === 'string' ? { message: options } : options
  // Errors/warnings get more time on screen — they're often longer or more important than a quick "Saved" pulse.
  const defaultDuration = next.severity === 'error' || next.severity === 'warning' ? 5000 : 3000
  const duration = next.autoHideDuration ?? defaultDuration

  switch (next.severity) {
    case 'success':
      toast.success(next.message, { duration })
      break
    case 'warning':
      toast.warning(next.message, { duration })
      break
    case 'error':
      toast.error(next.message, { duration })
      break
    default:
      toast(next.message, { duration })
  }
}

export function AppSnackbarProvider({ children }: { children: ReactNode }) {
  const showToast = useCallback((options: ToastOptions | string) => {
    showToastMessage(options)
  }, [])

  const value = useMemo(() => ({ showToast }), [showToast])

  return <AppSnackbarContext.Provider value={value}>{children}</AppSnackbarContext.Provider>
}
