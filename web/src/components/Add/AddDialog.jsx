import { useCallback, useMemo, useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import { torrentsHost, torrentUploadHost } from 'utils/Hosts'
import axios from 'axios'
import { useTranslation } from 'react-i18next'
import { NoImageIcon, AddItemIcon, TorrentIcon } from 'icons'
import debounce from 'lodash/debounce'
import { v4 as uuidv4 } from 'uuid'
import useChangeLanguage from 'utils/useChangeLanguage'
import { Cancel as CancelIcon } from '@material-ui/icons'
import { useDropzone } from 'react-dropzone'
import { useMediaQuery } from '@material-ui/core'

import {
  ButtonWrapper,
  CancelIconWrapper,
  ClearPosterButton,
  Content,
  Header,
  IconWrapper,
  LeftSide,
  Poster,
  PosterSuggestions,
  PosterSuggestionsItem,
  PosterWrapper,
  RightSide,
  RightSideBottomSectionFileSelected,
  RightSideBottomSectionNoFile,
  RightSideTopSection,
  TorrentIconWrapper,
} from './style'
import { checkImageURL, getMoviePosters } from './helpers'

export default function AddDialog({ handleClose }) {
  const { t } = useTranslation()
  const [torrentSource, setTorrentSource] = useState('')
  const [torrentSourceSelected, setTorrentSourceSelected] = useState(false)
  const [title, setTitle] = useState('')
  const [posterUrl, setPosterUrl] = useState('')
  const [isPosterUrlCorrect, setIsPosterUrlCorrect] = useState(false)
  const [posterList, setPosterList] = useState()
  const [isUserInteractedWithPoster, setIsUserInteractedWithPoster] = useState(false)
  const [currentLang] = useChangeLanguage()
  const [selectedFile, setSelectedFile] = useState()

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  const handleCapture = useCallback(files => {
    const [file] = files
    if (!file) return

    setSelectedFile(file)
    setTorrentSource(file.name)
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop: handleCapture, accept: '.torrent' })

  const removePoster = () => {
    setIsPosterUrlCorrect(false)
    setPosterUrl('')
  }

  const delayedPosterSearch = useMemo(
    () =>
      debounce(movieName => {
        getMoviePosters(movieName, currentLang === 'ru' ? 'ru' : 'en').then(urlList => {
          if (urlList) {
            setPosterList(urlList)
            if (isUserInteractedWithPoster) return

            const [firstPoster] = urlList
            checkImageURL(firstPoster).then(correctImage => {
              if (correctImage) {
                setIsPosterUrlCorrect(true)
                setPosterUrl(firstPoster)
              } else removePoster()
            })
          } else {
            setPosterList()
            if (isUserInteractedWithPoster) return

            removePoster()
          }
        })
      }, 700),
    [isUserInteractedWithPoster, currentLang],
  )

  const handleTorrentSourceChange = ({ target: { value } }) => setTorrentSource(value)
  const handleTitleChange = ({ target: { value } }) => {
    setTitle(value)
    delayedPosterSearch(value)
  }
  const handlePosterUrlChange = ({ target: { value } }) => {
    setPosterUrl(value)
    checkImageURL(value).then(setIsPosterUrlCorrect)
    setIsUserInteractedWithPoster(!!value)
    setPosterList()
  }

  const handleSave = () => {
    if (selectedFile) {
      // file save
      const data = new FormData()
      data.append('save', 'true')
      data.append('file', selectedFile)
      title && data.append('title', title)
      posterUrl && data.append('poster', posterUrl)
      axios.post(torrentUploadHost(), data).finally(handleClose)
    } else {
      // link save
      axios
        .post(torrentsHost(), { action: 'add', link: torrentSource, title, poster: posterUrl, save_to_db: true })
        .finally(handleClose)
    }
  }

  const clearSelectedFile = () => {
    setSelectedFile()
    setTorrentSource('')
  }

  const userChangesPosterUrl = url => {
    setPosterUrl(url)
    checkImageURL(url).then(setIsPosterUrlCorrect)
    setIsUserInteractedWithPoster(true)
  }

  return (
    <Dialog
      open
      onClose={handleClose}
      aria-labelledby='form-dialog-title'
      fullScreen={fullScreen}
      fullWidth
      maxWidth='md'
    >
      <Header>{t('AddNewTorrent')}</Header>

      <Content>
        <LeftSide>
          <TextField
            onChange={handleTitleChange}
            value={title}
            margin='dense'
            label={t('Title')}
            type='text'
            fullWidth
          />
          <TextField
            onChange={handlePosterUrlChange}
            value={posterUrl}
            margin='dense'
            label={t('AddPosterLinkInput')}
            type='url'
            fullWidth
          />

          <PosterWrapper>
            <Poster poster={+isPosterUrlCorrect}>
              {isPosterUrlCorrect ? <img src={posterUrl} alt='poster' /> : <NoImageIcon />}
            </Poster>

            <PosterSuggestions>
              {posterList
                ?.filter(url => url !== posterUrl)
                .slice(0, 12)
                .map(url => (
                  <PosterSuggestionsItem onClick={() => userChangesPosterUrl(url)} key={uuidv4()}>
                    <img src={url} alt='poster' />
                  </PosterSuggestionsItem>
                ))}
            </PosterSuggestions>

            <ClearPosterButton
              showbutton={+isPosterUrlCorrect}
              onClick={() => {
                removePoster()
                setIsUserInteractedWithPoster(true)
              }}
              color='primary'
              variant='contained'
              size='small'
              disabled={!posterUrl}
            >
              {t('Clear')}
            </ClearPosterButton>
          </PosterWrapper>
        </LeftSide>

        <RightSide>
          <RightSideTopSection active={torrentSourceSelected}>
            <TextField
              onChange={handleTorrentSourceChange}
              value={torrentSource}
              margin='dense'
              label={t('TorrentSourceLink')}
              helperText={t('TorrentSourceOptions')}
              type='text'
              fullWidth
              onFocus={() => setTorrentSourceSelected(true)}
              onBlur={() => setTorrentSourceSelected(false)}
              inputProps={{ autoComplete: 'off' }}
              disabled={!!selectedFile}
            />
          </RightSideTopSection>

          {selectedFile ? (
            <RightSideBottomSectionFileSelected>
              <TorrentIconWrapper>
                <TorrentIcon />

                <CancelIconWrapper onClick={clearSelectedFile}>
                  <CancelIcon />
                </CancelIconWrapper>
              </TorrentIconWrapper>
            </RightSideBottomSectionFileSelected>
          ) : (
            <RightSideBottomSectionNoFile isDragActive={isDragActive} {...getRootProps()}>
              <input {...getInputProps()} />
              <div>{t('AppendFile.Or')}</div>

              <IconWrapper>
                <AddItemIcon color='primary' />
                <div>{t('AppendFile.ClickOrDrag')}</div>
              </IconWrapper>
            </RightSideBottomSectionNoFile>
          )}
        </RightSide>
      </Content>

      <ButtonWrapper>
        <Button onClick={handleClose} color='primary' variant='outlined'>
          {t('Cancel')}
        </Button>

        <Button variant='contained' disabled={!torrentSource} onClick={handleSave} color='primary'>
          {t('Add')}
        </Button>
      </ButtonWrapper>
    </Dialog>
  )
}
