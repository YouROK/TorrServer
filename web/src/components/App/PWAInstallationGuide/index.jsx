import IconButton from '@material-ui/core/IconButton'
import CloseIcon from '@material-ui/icons/Close'
import { useState } from 'react'

import IOSShareIcon from './IOSShareIcon'
import { StyledWrapper, StyledHeader, StyledContent } from './style'

export function PWAInstallationGuide() {
  const [isOpen, setIsOpen] = useState(!JSON.parse(localStorage.getItem('pwaNotificationIsClosed')))

  return (
    <StyledWrapper isOpen={isOpen}>
      <StyledHeader>
        <img src='/apple-icon-180.png' width={50} alt='ts-icon' />
        Install application
        <IconButton
          size='small'
          aria-label='close'
          color='inherit'
          onClick={() => {
            setIsOpen(false)
            localStorage.setItem('pwaNotificationIsClosed', true)
          }}
        >
          <CloseIcon fontSize='small' />
        </IconButton>
      </StyledHeader>

      <StyledContent>
        <p>Install the app on your device to easily access it anytime. No app store. No download.</p>

        <p>VLC button will be added to open video instantly on the phone</p>

        <p>
          1. Tap on <IOSShareIcon />
        </p>

        <p>
          2. Select <span>Add to Home Screen</span>
        </p>
      </StyledContent>
    </StyledWrapper>
  )
}
