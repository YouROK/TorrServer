import { useState } from 'react'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import LibraryAddIcon from '@material-ui/icons/LibraryAdd'
import ListItemText from '@material-ui/core/ListItemText'
import { useTranslation } from 'react-i18next'
import { StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import { isStandaloneApp } from 'utils/Utils'

import AddDialog from './AddDialog'
import { StyledPWAAddButton } from './style'

export default function AddDialogButton({ isOffline, isLoading }) {
  const { t } = useTranslation()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const handleClickOpen = () => setIsDialogOpen(true)
  const handleClose = () => setIsDialogOpen(false)

  return (
    <div>
      <StyledMenuButtonWrapper disabled={isOffline || isLoading} button onClick={handleClickOpen}>
        {isStandaloneApp ? (
          <StyledPWAAddButton />
        ) : (
          <>
            <ListItemIcon>
              <LibraryAddIcon />
            </ListItemIcon>

            <ListItemText primary={t('AddFromLink')} />
          </>
        )}
      </StyledMenuButtonWrapper>

      {isDialogOpen && <AddDialog handleClose={handleClose} />}
    </div>
  )
}
