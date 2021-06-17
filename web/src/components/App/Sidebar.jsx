import Divider from '@material-ui/core/Divider'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { CreditCard as CreditCardIcon } from '@material-ui/icons'
import List from '@material-ui/core/List'
import { useTranslation } from 'react-i18next'
import AddDialogButton from 'components/Add'
import SettingsDialog from 'components/Settings'
import RemoveAll from 'components/RemoveAll'
import AboutDialog from 'components/About'
import CloseServer from 'components/CloseServer'

import { AppSidebarStyle } from './style'

export default function Sidebar({ isDrawerOpen, setIsDonationDialogOpen }) {
  const { t } = useTranslation()

  return (
    <AppSidebarStyle isDrawerOpen={isDrawerOpen}>
      <List>
        <AddDialogButton />
        <RemoveAll />
      </List>

      <Divider />

      <List>
        <SettingsDialog />
        <CloseServer />
      </List>

      <Divider />

      <List>
        <AboutDialog />

        <ListItem button onClick={() => setIsDonationDialogOpen(true)}>
          <ListItemIcon>
            <CreditCardIcon />
          </ListItemIcon>
          <ListItemText primary={t('Donate')} />
        </ListItem>
      </List>
    </AppSidebarStyle>
  )
}
