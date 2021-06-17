import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import AddDialog from '../Add/AddDialog'
import IconWrapper from './style'

export default function AddFirstTorrent() {
  const { t } = useTranslation()
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const handleClickOpen = () => setIsDialogOpen(true)
  const handleClose = () => setIsDialogOpen(false)

  return (
    <>
      <IconWrapper onClick={() => handleClickOpen(true)} isButton>
        <lord-icon
          src='https://cdn.lordicon.com/bbnkwdur.json'
          trigger='loop'
          colors='primary:#121331,secondary:#00A572'
          stroke='26'
          scale='60'
        />
        <div className='icon-label'>{t('NoTorrentsAdded')}</div>
      </IconWrapper>

      {isDialogOpen && <AddDialog handleClose={handleClose} />}
    </>
  )
}
