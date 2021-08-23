import { useTranslation } from 'react-i18next'
import { Checkbox, FormControlLabel } from '@material-ui/core'
import { useState } from 'react'

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
  const [isSnakeDebugMode, setIsSnakeDebugMode] = useState(
    JSON.parse(localStorage.getItem('isSnakeDebugMode')) || false,
  )

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
        <SectionTitle color='#000' mb={20}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>{t('Cache')}</span>

            <FormControlLabel
              control={
                <Checkbox
                  color='primary'
                  checked={isSnakeDebugMode}
                  disableRipple
                  onChange={({ target: { checked } }) => {
                    setIsSnakeDebugMode(checked)
                    localStorage.setItem('isSnakeDebugMode', checked)
                  }}
                />
              }
              label={t('DebugMode')}
              labelPlacement='start'
            />
          </div>
        </SectionTitle>

        <TorrentCache cache={cache} isSnakeDebugMode={isSnakeDebugMode} />
      </DetailedViewCacheSection>
    </>
  )
}
