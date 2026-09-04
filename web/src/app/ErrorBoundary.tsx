import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Button } from '@heroui/react'
import i18n from 'shared/i18n'

interface ErrorBoundaryProps {
  children: ReactNode
}

interface ErrorBoundaryState {
  hasError: boolean
  message?: string
}

/** Top-level crash guard — renders even if a provider below it throws. */
export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, message: error?.message || i18n.t('UnexpectedError') }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('UI error boundary caught:', error, info.componentStack)
  }

  handleReload = () => {
    window.location.reload()
  }

  render() {
    if (!this.state.hasError) return this.props.children

    return (
      <div className='grid min-h-dvh place-items-center gap-4 bg-background p-6 text-center'>
        <h1 className='m-0 text-2xl font-semibold text-foreground'>{i18n.t('SomethingWentWrong')}</h1>
        {this.state.message ? <p className='m-0 max-w-md text-sm text-muted'>{this.state.message}</p> : null}
        <Button variant='primary' onPress={this.handleReload}>
          {i18n.t('Reload')}
        </Button>
      </div>
    )
  }
}
