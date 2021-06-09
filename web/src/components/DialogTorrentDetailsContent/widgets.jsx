import {
  ArrowDownward as ArrowDownwardIcon,
  ArrowUpward as ArrowUpwardIcon,
  SwapVerticalCircle as SwapVerticalCircleIcon,
  ViewAgenda as ViewAgendaIcon,
  Widgets as WidgetsIcon,
  PhotoSizeSelectSmall as PhotoSizeSelectSmallIcon,
  Build as BuildIcon,
} from '@material-ui/icons'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { useTranslation } from 'react-i18next'
import { GETTING_INFO, IN_DB, CLOSED, PRELOAD, WORKING } from 'torrentStates'

import StatisticsField from './StatisticsField'

export const DownlodSpeedWidget = ({ data }) => {
  const { t } = useTranslation()
  return (
    <StatisticsField
      title={t('DownloadSpeed')}
      value={humanizeSize(data) || '0 B'}
      iconBg='#118f00'
      valueBg='#13a300'
      icon={ArrowDownwardIcon}
    />
  )
}

export const UploadSpeedWidget = ({ data }) => {
  const { t } = useTranslation()
  return (
    <StatisticsField
      title={t('UploadSpeed')}
      value={humanizeSize(data) || '0 B'}
      iconBg='#0146ad'
      valueBg='#0058db'
      icon={ArrowUpwardIcon}
    />
  )
}

export const PeersWidget = ({ data }) => {
  const { t } = useTranslation()
  return (
    <StatisticsField
      title={t('Peers')}
      value={getPeerString(data) || '[0] 0 / 0'}
      iconBg='#cdc118'
      valueBg='#d8cb18'
      icon={SwapVerticalCircleIcon}
    />
  )
}

export const PiecesCountWidget = ({ data }) => {
  const { t } = useTranslation()
  return <StatisticsField title={t('PiecesCount')} value={data} iconBg='#b6c95e' valueBg='#c0d076' icon={WidgetsIcon} />
}

export const PiecesLengthWidget = ({ data }) => {
  const { t } = useTranslation()
  return (
    <StatisticsField
      title={t('PiecesLength')}
      value={humanizeSize(data)}
      iconBg='#0982c8'
      valueBg='#098cd7'
      icon={PhotoSizeSelectSmallIcon}
    />
  )
}

export const StatusWidget = ({ stat }) => {
  const { t } = useTranslation()

  const values = {
    [GETTING_INFO]: t('TorrentGettingInfo'),
    [PRELOAD]: t('TorrentPreload'),
    [WORKING]: t('TorrentWorking'),
    [CLOSED]: t('TorrentClosed'),
    [IN_DB]: t('TorrentInDb'),
  }

  return (
    <StatisticsField
      title={t('TorrentStatus')}
      value={values[stat]}
      iconBg='#aea25b'
      valueBg='#b4aa6e'
      icon={BuildIcon}
    />
  )
}

export const SizeWidget = ({ data }) => {
  const { t } = useTranslation()
  return (
    <StatisticsField
      title={t('TorrentSize')}
      value={humanizeSize(data)}
      iconBg='#9b01ad'
      valueBg='#ac03bf'
      icon={ViewAgendaIcon}
    />
  )
}
