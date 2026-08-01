import { FormControl, MenuItem, Select, Tooltip, useMediaQuery } from '@material-ui/core'
import { Tv as TvIcon } from '@material-ui/icons'
import { useTranslation } from 'react-i18next'
import { PLAYBACK_ROUTING_MODES, usePlayback } from 'playback/PlaybackContext'
import { useBooleanPlayerPreference } from 'utils/PlayerPreferences'

const PlaybackTargetSelector = () => {
  const { t } = useTranslation()
  const isCompact = useMediaQuery('(max-width:700px)')
  const {
    isLoadingDevices,
    remotePlaybackEnabled,
    routingMode,
    selectedTarget,
    setTargetId,
    showTargetSelector,
    targetId,
    targets,
  } = usePlayback()
  const isVlcUsed = useBooleanPlayerPreference('isVlcUsed')

  if (!isVlcUsed || !remotePlaybackEnabled || routingMode === PLAYBACK_ROUTING_MODES.LOCAL) return null

  const label = selectedTarget?.name || t('Playback.NoDevice')
  const labelContent = (
    <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'inherit' }}>
      <TvIcon fontSize='small' />
      <span
        style={{
          maxWidth: isCompact ? 75 : 165,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }}
      >
        {label}
      </span>
    </span>
  )

  if (!showTargetSelector) {
    return (
      <Tooltip title={t('Playback.PrimaryDeviceActive')}>
        <span style={{ fontSize: isCompact ? '0.75rem' : '0.875rem' }}>{labelContent}</span>
      </Tooltip>
    )
  }

  return (
    <Tooltip title={t('Playback.PlayOn')}>
      <FormControl style={{ minWidth: isCompact ? 90 : 150, maxWidth: isCompact ? 120 : 220 }}>
        <Select
          disableUnderline
          displayEmpty
          value={targetId}
          disabled={isLoadingDevices || targets.length === 0}
          onChange={({ target: { value } }) => setTargetId(value)}
          inputProps={{ 'aria-label': t('Playback.PlayOn') }}
          renderValue={() => labelContent}
          style={{ color: 'inherit', fontSize: isCompact ? '0.75rem' : '0.875rem' }}
        >
          {targets.length === 0 && (
            <MenuItem value='' disabled>
              {t('Playback.NoDevice')}
            </MenuItem>
          )}
          {targets.map(target => (
            <MenuItem key={target.id} value={target.id}>
              {target.name}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Tooltip>
  )
}

export default PlaybackTargetSelector
