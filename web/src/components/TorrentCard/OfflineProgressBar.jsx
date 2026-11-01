import { memo } from 'react'
import { useTranslation } from 'react-i18next'
import { humanizeSize } from 'utils/Utils'
import { getOfflineFraction, getOfflineLoadedSize, OFFLINE_COLORS } from 'utils/offline'

// Thin progress line on the bottom edge of a torrent card. Put it as the LAST child of the card root,
// which must be `position: relative` (the bar is absolutely positioned). Renders nothing for
// torrents that are not in the offline mode.
const OfflineProgressBar = ({ torrent }) => {
  const { t } = useTranslation()
  const state = torrent?.offline_state
  if (!state) return null
  const percent = getOfflineFraction(torrent)

  return (
    <div
      role='progressbar'
      aria-label={t('OfflineStorage', 'Offline storage')}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={Math.floor(percent)}
      style={{
        position: 'absolute',
        left: 0,
        right: 0,
        bottom: 0,
        height: 4,
        overflow: 'hidden',
        pointerEvents: 'none',
        background: 'rgba(128, 128, 128, 0.25)',
        // follow the rounded corners of the card, if it has any
        borderBottomLeftRadius: 'inherit',
        borderBottomRightRadius: 'inherit',
      }}
    >
      <div
        style={{
          width: `${percent}%`,
          height: '100%',
          background: OFFLINE_COLORS[state],
          transition: 'width 0.6s linear',
        }}
      />
    </div>
  )
}

// "downloaded / total" for the Size field of an offline torrent, `t` is the i18n function.
// The unit is not repeated when it is the same: "1.2 / 4.5 GB" (shorter - the size cell is narrow).
export const getOfflineSizeLabel = (torrent, t) => {
  const zero = `0 ${t('B')}`
  const loaded = humanizeSize(getOfflineLoadedSize(torrent)) || zero
  const total = humanizeSize(torrent.torrent_size) || zero
  const [loadedValue, loadedUnit] = loaded.split(' ')
  const totalUnit = total.split(' ')[1]
  return `${loadedUnit === totalUnit ? loadedValue : loaded} / ${total}`
}

export default memo(OfflineProgressBar)
