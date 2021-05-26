import { useState } from 'react'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import LibraryAddIcon from '@material-ui/icons/LibraryAdd'
import ListItemText from '@material-ui/core/ListItemText'
import ListItem from '@material-ui/core/ListItem'

import AddDialog from './AddDialog'

export default function AddDialogButton() {
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const handleClickOpen = () => setIsDialogOpen(true)
  const handleClose = () => setIsDialogOpen(false)

  return (
    <div>
      <ListItem button key='Add' onClick={handleClickOpen}>
        <ListItemIcon>
          <LibraryAddIcon />
        </ListItemIcon>
        <ListItemText primary='Add from link' />
      </ListItem>

      {isDialogOpen && <AddDialog handleClose={handleClose} />}
    </div>
  )
}
