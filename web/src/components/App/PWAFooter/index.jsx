import { CreditCard as CreditCardIcon } from '@material-ui/icons'
import { useTranslation } from 'react-i18next'
import CloseServer from 'components/CloseServer'
import { StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import AddDialogButton from 'components/Add'
import AboutDialog from 'components/About'
import SettingsDialogButton from 'components/Settings'

import StyledPWAFooter from './style'

export default function PWAFooter({ setIsDonationDialogOpen, isOffline, isLoading }) {
  const { t } = useTranslation()

  return (
    <StyledPWAFooter>
      <CloseServer isOffline={isOffline} isLoading={isLoading} />

      <StyledMenuButtonWrapper onClick={() => setIsDonationDialogOpen(true)}>
        <CreditCardIcon />

        <div>{t('Donate')}</div>
      </StyledMenuButtonWrapper>

      <AddDialogButton isOffline={isOffline} isLoading={isLoading} />

      <AboutDialog />

      <SettingsDialogButton isOffline={isOffline} isLoading={isLoading} />
    </StyledPWAFooter>
  )
}
