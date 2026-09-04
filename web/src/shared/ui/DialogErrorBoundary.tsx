import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Button } from '@heroui/react'
import i18n from 'shared/i18n'

interface DialogErrorBoundaryProps {
  children: ReactNode
  onClose?: () => void
}

interface DialogErrorBoundaryState {
  hasError: boolean
  message?: string
}

/** Isolates a single dialog crash so the rest of the shell stays usable. */
export default class DialogErrorBoundary extends Component<DialogErrorBoundaryProps, DialogErrorBoundaryState> {
  state: DialogErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError(error: Error): DialogErrorBoundaryState {
    return { hasError: true, message: error?.message || i18n.t('UnexpectedError') }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Dialog error boundary caught:', error, info.componentStack)
  }

  handleDismiss = () => {
    this.setState({ hasError: false, message: undefined })
    this.props.onClose?.()
  }

  render() {
    if (!this.state.hasError) return this.props.children

    return (
      <div className='fixed inset-0 z-[100] grid place-items-center bg-black/50 p-6'>
        <div className='max-w-md rounded-2xl border border-border bg-surface p-6 text-center shadow-xl'>
          <h2 className='m-0 text-lg font-semibold text-foreground'>{i18n.t('SomethingWentWrong')}</h2>
          {this.state.message ? <p className='mt-2 text-sm text-muted'>{this.state.message}</p> : null}
          <Button className='mt-4' variant='primary' onPress={this.handleDismiss}>
            {i18n.t('Close')}
          </Button>
        </div>
      </div>
    )
  }
}
