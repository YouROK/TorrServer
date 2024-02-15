import { Button } from '@material-ui/core';
import { useEffect, useRef, useState } from 'react';

export default function UnsafeButton({ timeout, children, disabled, ...props }) {
    const [timeLeft, setTimeLeft] = useState(timeout || 7)
    const [buttonDisabled, setButtonDisabled] = useState(disabled || timeLeft > 0)
    const handleTimerTick = () => {
        const newTimeLeft = timeLeft - 1
        setTimeLeft(newTimeLeft)
        if (newTimeLeft <= 0) {
            setButtonDisabled(disabled)
        }
    }
    const getTimerText = () => !disabled && timeLeft > 0 ? ` (${timeLeft})` : ''
    useEffect(() => {
        if (disabled || !timeLeft) { return }
        const intervalId = setInterval(handleTimerTick, 1000)
        return () => clearInterval(intervalId)
    }, [timeLeft])

    return (
        <Button disabled={buttonDisabled} {...props}>
            {children} {getTimerText()}
        </Button>
    )
}
