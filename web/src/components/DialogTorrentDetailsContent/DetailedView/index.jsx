import { useTranslation } from 'react-i18next'

import { SectionTitle, WidgetWrapper } from '../style'
import { DetailedViewCacheSection, DetailedViewWidgetSection } from './style'
import TorrentCache from '../TorrentCache'
import {
  SizeWidget,
  PiecesLengthWidget,
  StatusWidget,
  PiecesCountWidget,
  PeersWidget,
  UploadSpeedWidget,
  DownlodSpeedWidget,
} from '../widgets'

export default function DetailedView({
  downloadSpeed,
  uploadSpeed,
  torrent,
  torrentSize,
  PiecesCount,
  PiecesLength,
  stat,
  cache,
}) {
  const { t } = useTranslation()

  return (
    <>
      <DetailedViewWidgetSection>
        <SectionTitle mb={20}>{t('Data')}</SectionTitle>
        <WidgetWrapper detailedView>
          <DownlodSpeedWidget data={downloadSpeed} />
          <UploadSpeedWidget data={uploadSpeed} />
          <PeersWidget data={torrent} />
          <SizeWidget data={torrentSize} />
          <PiecesCountWidget data={PiecesCount} />
          <PiecesLengthWidget data={PiecesLength} />
          <StatusWidget stat={stat} />
        </WidgetWrapper>
      </DetailedViewWidgetSection>

      <DetailedViewCacheSection>
        <SectionTitle mb={20}>{t('Cache')}</SectionTitle>
        <TorrentCache cache={cache} />
      </DetailedViewCacheSection>
    </>
  )
}
