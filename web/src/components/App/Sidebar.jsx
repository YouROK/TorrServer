import Divider from '@material-ui/core/Divider'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { CreditCard as CreditCardIcon, Language as LanguageIcon } from '@material-ui/icons'
import List from '@material-ui/core/List'
import { useTranslation } from 'react-i18next'
import useChangeLanguage from 'utils/useChangeLanguage'
import AddDialogButton from 'components/Add'
import SettingsDialog from 'components/Settings'
import RemoveAll from 'components/RemoveAll'
import AboutDialog from 'components/About'
import CloseServer from 'components/CloseServer'

import { AppSidebarStyle } from './style'

export default function Sidebar({ isDrawerOpen, setIsDonationDialogOpen }) {
  const [currentLang, changeLang] = useChangeLanguage()
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
        <ListItem button onClick={() => (currentLang === 'en' ? changeLang('ru') : changeLang('en'))}>
          <ListItemIcon>
            <LanguageIcon />
          </ListItemIcon>
          <ListItemText primary={t('ChooseLanguage')} />
        </ListItem>
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
