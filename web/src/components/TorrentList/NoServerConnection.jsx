import { useTheme } from '@material-ui/core'
import { useTranslation } from 'react-i18next'

import IconWrapper from './style'

export default function NoServerConnection() {
  const { t } = useTranslation()
  const primary = useTheme().palette.primary.main

  return (
    <IconWrapper>
      <lord-icon
        src='https://cdn.lordicon.com/wrprwmwt.json'
        trigger='loop'
        colors={`primary:#121331,secondary:${primary}`}
        stroke='26'
        scale='60'
      />
      <div className='icon-label'>{t('Offline')}</div>
    </IconWrapper>
  )
}
