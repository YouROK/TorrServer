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

import StatisticsField from './StatisticsField'

export const DownlodSpeedWidget = ({ data }) => (
  <StatisticsField
    title='Download speed'
    value={humanizeSize(data) || '0 B'}
    iconBg='#118f00'
    valueBg='#13a300'
    icon={ArrowDownwardIcon}
  />
)

export const UploadSpeedWidget = ({ data }) => (
  <StatisticsField
    title='Upload speed'
    value={humanizeSize(data) || '0 B'}
    iconBg='#0146ad'
    valueBg='#0058db'
    icon={ArrowUpwardIcon}
  />
)

export const PeersWidget = ({ data }) => (
  <StatisticsField
    title='Peers'
    value={getPeerString(data)}
    iconBg='#cdc118'
    valueBg='#d8cb18'
    icon={SwapVerticalCircleIcon}
  />
)

export const PiecesCountWidget = ({ data }) => (
  <StatisticsField title='Pieces count' value={data} iconBg='#b6c95e' valueBg='#c0d076' icon={WidgetsIcon} />
)
export const PiecesLengthWidget = ({ data }) => (
  <StatisticsField
    title='Pieces length'
    value={humanizeSize(data)}
    iconBg='#0982c8'
    valueBg='#098cd7'
    icon={PhotoSizeSelectSmallIcon}
  />
)
export const StatusWidget = ({ data }) => (
  <StatisticsField title='Torrent status' value={data} iconBg='#aea25b' valueBg='#b4aa6e' icon={BuildIcon} />
)

export const SizeWidget = ({ data }) => (
  <StatisticsField
    title='Torrent size'
    value={humanizeSize(data)}
    iconBg='#9b01ad'
    valueBg='#ac03bf'
    icon={ViewAgendaIcon}
  />
)
