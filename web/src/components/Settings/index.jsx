import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useState } from 'react'
import SettingsIcon from '@material-ui/icons/Settings'
import { useTranslation } from 'react-i18next'
import { StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import { isStandaloneApp } from 'utils/Utils'

import SettingsDialog from './SettingsDialog'

export default function SettingsDialogButton({ isOffline, isLoading }) {
  const { t } = useTranslation()
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const handleClickOpen = () => setIsDialogOpen(true)
  const handleClose = () => setIsDialogOpen(false)

  return (
    <div>
      <StyledMenuButtonWrapper disabled={isOffline || isLoading} button onClick={handleClickOpen}>
        {isStandaloneApp ? (
          <>
            <SettingsIcon />
            <div>{t('SettingsDialog.Settings')}</div>
          </>
        ) : (
          <>
            <ListItemIcon>
              <SettingsIcon />
            </ListItemIcon>

            <ListItemText primary={t('SettingsDialog.Settings')} />
          </>
        )}
      </StyledMenuButtonWrapper>

      {isDialogOpen && <SettingsDialog handleClose={handleClose} />}
    </div>
  )
}
