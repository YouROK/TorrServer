import {
  ArrowDownward as ArrowDownwardIcon,
  ArrowUpward as ArrowUpwardIcon,
  SwapVerticalCircle as SwapVerticalCircleIcon,
  ViewAgenda as ViewAgendaIcon,
  Widgets as WidgetsIcon,
  PhotoSizeSelectSmall as PhotoSizeSelectSmallIcon,
  Build as BuildIcon,
  Category as CategoryIcon,
} from '@material-ui/icons'
import { getPeerString, humanizeSize, humanizeSpeed } from 'utils/Utils'
import { useTranslation } from 'react-i18next'
import { GETTING_INFO, IN_DB, CLOSED, PRELOAD, WORKING } from 'torrentStates'
import { TORRENT_CATEGORIES } from 'components/categories'

import StatisticsField from '../StatisticsField'
import useGetWidgetColors from './useGetWidgetColors'

export const DownlodSpeedWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('downloadSpeed')

  return (
    <StatisticsField
      title={t('DownloadSpeed')}
      value={humanizeSpeed(data) || `0 ${t('bps')}`}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
      icon={ArrowDownwardIcon}
    />
  )
}

export const UploadSpeedWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('uploadSpeed')

  return (
    <StatisticsField
      title={t('UploadSpeed')}
      value={humanizeSpeed(data) || `0 ${t('bps')}`}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
      icon={ArrowUpwardIcon}
    />
  )
}

export const PeersWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('peers')

  return (
    <StatisticsField
      title={t('Peers')}
      value={getPeerString(data) || '0 / 0 · 0'}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
      icon={SwapVerticalCircleIcon}
    />
  )
}

export const PiecesCountWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('piecesCount')

  return (
    <StatisticsField
      title={t('PiecesCount')}
      value={data}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
      icon={WidgetsIcon}
    />
  )
}

export const PiecesLengthWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('piecesLength')

  return (
    <StatisticsField
      title={t('PiecesLength')}
      value={humanizeSize(data)}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
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
  const { iconBGColor, valueBGColor } = useGetWidgetColors('status')

  return (
    <StatisticsField
      title={t('TorrentStatus')}
      value={values[stat]}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
      icon={BuildIcon}
    />
  )
}

export const SizeWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('size')

  return (
    <StatisticsField
      title={t('TorrentSize')}
      value={humanizeSize(data)}
      iconBg={iconBGColor}
      valueBg={valueBGColor}
      icon={ViewAgendaIcon}
    />
  )
}

export const CategoryWidget = ({ data }) => {
  const { t } = useTranslation()
  const { iconBGColor, valueBGColor } = useGetWidgetColors('category')
  // main categories
  const catIndex = TORRENT_CATEGORIES.findIndex(e => e.key === data)
  const catArray = TORRENT_CATEGORIES.find(e => e.key === data)

  if (data) {
    return (
      <StatisticsField
        title={t('Category')}
        value={catIndex >= 0 ? t(catArray.name) : data.length > 1 ? data.charAt(0).toUpperCase() + data.slice(1) : data}
        iconBg={iconBGColor}
        valueBg={valueBGColor}
        icon={CategoryIcon}
      />
    )
  }

  return null
}
