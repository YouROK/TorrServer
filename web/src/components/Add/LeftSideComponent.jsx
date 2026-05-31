import { useTranslation } from 'react-i18next'
import { useDropzone } from 'react-dropzone'
import { AddItemIcon } from 'icons'
import TextField from '@material-ui/core/TextField'
import { useState } from 'react'

import { IconWrapper, LeftSide, LeftSideBottomSectionNoFile, LeftSideTopSection } from './style'

export default function LeftSideComponent({
  setIsUserInteractedWithPoster,
  setSelectedFile,
  torrentSource,
  setTorrentSource,
}) {
  const { t } = useTranslation()

  const handleCapture = files => {
    if (!files.length) return
    setIsUserInteractedWithPoster(false)
    setSelectedFile(files.length === 1 ? files[0] : files)
  }

  const [isTorrentSourceActive, setIsTorrentSourceActive] = useState(false)
  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop: handleCapture,
    accept: '.torrent',
    multiple: true,
  })

  const handleTorrentSourceChange = ({ target: { value } }) => setTorrentSource(value)

  return (
    <LeftSide>
      <LeftSideTopSection active={isTorrentSourceActive}>
        <TextField
          onChange={handleTorrentSourceChange}
          value={torrentSource}
          margin='dense'
          label={t('AddDialog.TorrentSourceLink')}
          helperText={t('AddDialog.TorrentSourceOptions')}
          style={{ marginTop: '1em' }}
          type='text'
          fullWidth
          variant='outlined'
          onFocus={() => setIsTorrentSourceActive(true)}
          onBlur={() => setIsTorrentSourceActive(false)}
          inputProps={{ autoComplete: 'off' }}
        />
      </LeftSideTopSection>

      <LeftSideBottomSectionNoFile isDragActive={isDragActive} {...getRootProps()}>
        <input {...getInputProps()} />
        <div>{t('AddDialog.AppendFile.Or')}</div>

        <IconWrapper>
          <AddItemIcon color='primary' />
          <div>{t('AddDialog.AppendFile.ClickOrDrag')}</div>
        </IconWrapper>
      </LeftSideBottomSectionNoFile>
    </LeftSide>
  )
}
