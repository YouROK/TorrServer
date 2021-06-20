import { useTranslation } from 'react-i18next'
import { mainColors } from 'style/colors'

import IconWrapper from './style'

export default function NoServerConnection() {
  const { t } = useTranslation()

  return (
    <IconWrapper>
      <lord-icon
        src='https://cdn.lordicon.com/wrprwmwt.json'
        trigger='loop'
        colors={`primary:#121331,secondary:${mainColors.primary}`}
        stroke='26'
        scale='60'
      />
      <div className='icon-label'>{t('Offline')}</div>
    </IconWrapper>
  )
}
