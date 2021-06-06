import { playlistAllHost } from 'utils/Hosts'
import Divider from '@material-ui/core/Divider'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import AddDialogButton from 'components/Add'
import RemoveAll from 'components/RemoveAll'
import SettingsDialog from 'components/Settings'
import AboutDialog from 'components/About'
import UploadDialog from 'components/Upload'
import { CreditCard as CreditCardIcon, List as ListIcon } from '@material-ui/icons'
import List from '@material-ui/core/List'
import CloseServer from 'components/CloseServer'
import { useTranslation } from 'react-i18next'

import { AppSidebarStyle } from './style'

export default function Sidebar({ isDrawerOpen, setIsDonationDialogOpen }) {
  const { t } = useTranslation()
  return (
    <AppSidebarStyle isDrawerOpen={isDrawerOpen}>
      <List>
        <AddDialogButton />
        <UploadDialog />
        <RemoveAll />
        <ListItem button component='a' key={t('PlaylistAll')} target='_blank' href={playlistAllHost()}>
          <ListItemIcon>
            <ListIcon />
          </ListItemIcon>
          <ListItemText primary={t('PlaylistAll')} />
        </ListItem>
      </List>

      <Divider />

      <List>
        <SettingsDialog />
        <AboutDialog />
        <CloseServer />
      </List>

      <Divider />

      <List>
        <ListItem button key='Donation' onClick={() => setIsDonationDialogOpen(true)}>
          <ListItemIcon>
            <CreditCardIcon />
          </ListItemIcon>
          <ListItemText primary={t('Donate')} />
        </ListItem>
      </List>
    </AppSidebarStyle>
  )
}
