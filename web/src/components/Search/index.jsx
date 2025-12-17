import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import SearchIcon from '@material-ui/icons/Search'
import ListItemText from '@material-ui/core/ListItemText'
import ListItem from '@material-ui/core/ListItem'

import SearchDialog from './SearchDialog'

export default function SearchDialogButton({ isOffline, isLoading }) {
  const { t } = useTranslation()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const handleClickOpen = () => setIsDialogOpen(true)
  const handleClose = () => setIsDialogOpen(false)

  return (
    <>
      <ListItem button onClick={handleClickOpen} disabled={isOffline || isLoading}>
        <ListItemIcon>
          <SearchIcon />
        </ListItemIcon>
        <ListItemText primary={t('Search')} />
      </ListItem>

      {isDialogOpen && <SearchDialog handleClose={handleClose} />}
    </>
  )
}
