import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useState } from 'react'
import SettingsIcon from '@material-ui/icons/Settings'
import { useTranslation } from 'react-i18next'

import SettingsDialog from './SettingsDialog'

export default function SettingsDialogButton() {
  const { t } = useTranslation()
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const handleClickOpen = () => setIsDialogOpen(true)
  const handleClose = () => setIsDialogOpen(false)

  return (
    <div>
      <ListItem button onClick={handleClickOpen}>
        <ListItemIcon>
          <SettingsIcon />
        </ListItemIcon>
        <ListItemText primary={t('Settings')} />
      </ListItem>

      {isDialogOpen && <SettingsDialog handleClose={handleClose} />}
    </div>
  )
}
