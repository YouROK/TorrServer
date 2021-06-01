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

export default function Test({
  downloadSpeed,
  uploadSpeed,
  torrent,
  torrentSize,
  PiecesCount,
  PiecesLength,
  statString,
  cache,
}) {
  return (
    <>
      <DetailedViewWidgetSection>
        <SectionTitle mb={20}>Data</SectionTitle>
        <WidgetWrapper detailedView>
          <DownlodSpeedWidget data={downloadSpeed} />
          <UploadSpeedWidget data={uploadSpeed} />
          <PeersWidget data={torrent} />
          <SizeWidget data={torrentSize} />
          <PiecesCountWidget data={PiecesCount} />
          <PiecesLengthWidget data={PiecesLength} />
          <StatusWidget data={statString} />
        </WidgetWrapper>
      </DetailedViewWidgetSection>

      <DetailedViewCacheSection>
        <SectionTitle mb={20}>Cache</SectionTitle>
        <TorrentCache cache={cache} />
      </DetailedViewCacheSection>
    </>
  )
}
