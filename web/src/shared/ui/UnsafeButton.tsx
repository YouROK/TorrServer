import { Button } from '@heroui/react'
import { useEffect, useState, type ReactNode } from 'react'

interface UnsafeButtonProps {
  timeout?: number
  children?: ReactNode
  isDisabled?: boolean
  onPress?: () => void
  variant?: 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'danger'
  className?: string
}

/** Destructive action button with optional countdown delay. */
export default function UnsafeButton({
  timeout = 5,
  children,
  isDisabled,
  onPress,
  variant = 'danger',
  className,
}: UnsafeButtonProps) {
  const [startedAt, setStartedAt] = useState(() => Date.now())
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (isDisabled || timeout <= 0) return undefined
    const start = Date.now()
    // eslint-disable-next-line react-hooks/set-state-in-effect -- restart countdown when enabled/timeout changes
    setStartedAt(start)
    setNow(start)
    const intervalId = window.setInterval(() => setNow(Date.now()), 250)
    return () => window.clearInterval(intervalId)
  }, [isDisabled, timeout])

  const timeLeft = isDisabled || timeout <= 0 ? 0 : Math.max(0, timeout - Math.floor((now - startedAt) / 1000))
  const buttonDisabled = Boolean(isDisabled) || timeLeft > 0
  const timerText = !isDisabled && timeLeft > 0 ? ` (${timeLeft})` : ''

  return (
    <Button variant={variant} isDisabled={buttonDisabled} onPress={onPress} className={className}>
      {children}
      {timerText}
    </Button>
  )
}
