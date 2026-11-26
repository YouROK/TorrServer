import { CreditCard as CreditCardIcon, ViewList as ViewListIcon } from '@material-ui/icons'
import { useTranslation } from 'react-i18next'
import { StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import AddDialogButton from 'components/Add'
import AboutDialog from 'components/About'
import SettingsDialogButton from 'components/Settings'

import StyledPWAFooter from './style'

export default function PWAFooter({ setIsDonationDialogOpen, onBackToList, isOffline, isLoading }) {
  const { t } = useTranslation()

  return (
    <StyledPWAFooter>
      <StyledMenuButtonWrapper onClick={onBackToList}>
        <ViewListIcon />

        <div>{t('Torrents')}</div>
      </StyledMenuButtonWrapper>

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
