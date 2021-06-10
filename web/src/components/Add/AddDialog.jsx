import { useEffect, useMemo, useState } from 'react'
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
import parseTorrent from 'parse-torrent'
import ptt from 'parse-torrent-title'
import CircularProgress from '@material-ui/core/CircularProgress'
import usePreviousState from 'utils/usePreviousState'

import { checkImageURL, getMoviePosters, chechTorrentSource } from './helpers'
import {
  ButtonWrapper,
  CancelIconWrapper,
  ClearPosterButton,
  PosterLanguageSwitch,
  Content,
  Header,
  IconWrapper,
  RightSide,
  Poster,
  PosterSuggestions,
  PosterSuggestionsItem,
  PosterWrapper,
  LeftSide,
  LeftSideBottomSectionFileSelected,
  LeftSideBottomSectionNoFile,
  LeftSideTopSection,
  TorrentIconWrapper,
  RightSideContainer,
} from './style'

const parseTorrentTitle = (parsingSource, callback) => {
  parseTorrent.remote(parsingSource, (err, { name, files } = {}) => {
    if (!name || err) return callback(null)

    const torrentName = ptt.parse(name).title
    const nameOfFileInsideTorrent = files ? ptt.parse(files[0].name).title : null

    let newTitle = torrentName
    if (nameOfFileInsideTorrent) {
      // taking shorter title because in most cases it is more accurate
      newTitle = torrentName.length < nameOfFileInsideTorrent.length ? torrentName : nameOfFileInsideTorrent
    }

    callback(newTitle)
  })
}

export default function AddDialog({ handleClose }) {
  const { t } = useTranslation()
  const [torrentSource, setTorrentSource] = useState('')
  const [isTorrentSourceActive, setIsTorrentSourceActive] = useState(false)
  const [title, setTitle] = useState('')
  const [posterUrl, setPosterUrl] = useState('')
  const [isPosterUrlCorrect, setIsPosterUrlCorrect] = useState(false)
  const [isTorrentSourceCorrect, setIsTorrentSourceCorrect] = useState(false)
  const [posterList, setPosterList] = useState()
  const [isUserInteractedWithPoster, setIsUserInteractedWithPoster] = useState(false)
  const [currentLang] = useChangeLanguage()
  const [selectedFile, setSelectedFile] = useState()
  const [posterSearchLanguage, setPosterSearchLanguage] = useState(currentLang === 'ru' ? 'ru' : 'en')
  const [isLoadingButton, setIsLoadingButton] = useState(false)
  const [skipDebounce, setSkipDebounce] = useState(false)

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  const posterSearch = useMemo(
    () =>
      (movieName, language, settings = {}) => {
        const { shouldRefreshMainPoster = false } = settings

        getMoviePosters(movieName, language).then(urlList => {
          if (urlList) {
            setPosterList(urlList)
            if (!shouldRefreshMainPoster && isUserInteractedWithPoster) return

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
      },
    [isUserInteractedWithPoster],
  )

  const delayedPosterSearch = useMemo(() => debounce(posterSearch, 2700), [posterSearch])

  const prevTitleState = usePreviousState(title)
  const prevTorrentSourceState = usePreviousState(torrentSource)

  useEffect(() => {
    // if torrentSource is updated then we are checking that source is valid and getting title from the source
    const torrentSourceChanged = torrentSource !== prevTorrentSourceState

    const isCorrectSource = chechTorrentSource(torrentSource)
    if (!isCorrectSource) return

    setIsTorrentSourceCorrect(true)

    if (torrentSourceChanged) {
      parseTorrentTitle(selectedFile || torrentSource, newTitle => {
        if (!newTitle) return

        setSkipDebounce(true)
        setTitle(newTitle)
      })
    }
  }, [prevTorrentSourceState, selectedFile, torrentSource])

  useEffect(() => {
    // if title exists and title was changed then search poster.
    const titleChanged = title !== prevTitleState
    if (!titleChanged) return

    if (skipDebounce) {
      posterSearch(title, posterSearchLanguage)
      setSkipDebounce(false)
    } else {
      delayedPosterSearch(title, posterSearchLanguage)
    }
  }, [title, prevTitleState, delayedPosterSearch, posterSearch, posterSearchLanguage, skipDebounce])

  useEffect(() => {
    if (!selectedFile && !torrentSource) {
      setTitle('')
      setPosterUrl('')
      setPosterList()
      setIsPosterUrlCorrect(false)
    }
  }, [selectedFile, torrentSource])

  const handleCapture = files => {
    const [file] = files
    if (!file) return

    setIsUserInteractedWithPoster(false)
    setSelectedFile(file)
    setTorrentSource(file.name)
  }

  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop: handleCapture, accept: '.torrent' })

  const removePoster = () => {
    setIsPosterUrlCorrect(false)
    setPosterUrl('')
  }

  const handleTorrentSourceChange = ({ target: { value } }) => setTorrentSource(value)
  const handleTitleChange = ({ target: { value } }) => setTitle(value)
  const handlePosterUrlChange = ({ target: { value } }) => {
    setPosterUrl(value)
    checkImageURL(value).then(setIsPosterUrlCorrect)
    setIsUserInteractedWithPoster(!!value)
    setPosterList()
  }

  const handleSave = () => {
    setIsLoadingButton(true)

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
          <LeftSideTopSection active={isTorrentSourceActive}>
            <TextField
              onChange={handleTorrentSourceChange}
              value={torrentSource}
              margin='dense'
              label={t('TorrentSourceLink')}
              helperText={t('TorrentSourceOptions')}
              type='text'
              fullWidth
              onFocus={() => setIsTorrentSourceActive(true)}
              onBlur={() => setIsTorrentSourceActive(false)}
              inputProps={{ autoComplete: 'off' }}
              disabled={!!selectedFile}
            />
          </LeftSideTopSection>

          {selectedFile ? (
            <LeftSideBottomSectionFileSelected>
              <TorrentIconWrapper>
                <TorrentIcon />

                <CancelIconWrapper onClick={clearSelectedFile}>
                  <CancelIcon />
                </CancelIconWrapper>
              </TorrentIconWrapper>
            </LeftSideBottomSectionFileSelected>
          ) : (
            <LeftSideBottomSectionNoFile isDragActive={isDragActive} {...getRootProps()}>
              <input {...getInputProps()} />
              <div>{t('AppendFile.Or')}</div>

              <IconWrapper>
                <AddItemIcon color='primary' />
                <div>{t('AppendFile.ClickOrDrag')}</div>
              </IconWrapper>
            </LeftSideBottomSectionNoFile>
          )}
        </LeftSide>

        <RightSide>
          {/* <RightSideContainer isHidden={!isTorrentSourceCorrect}> */}
          <RightSideContainer isHidden={false}>
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

              {currentLang !== 'en' && (
                <PosterLanguageSwitch
                  onClick={() => {
                    const newLanguage = posterSearchLanguage === 'en' ? 'ru' : 'en'
                    setPosterSearchLanguage(newLanguage)
                    posterSearch(title, newLanguage, { shouldRefreshMainPoster: true })
                    console.log(':334')
                  }}
                  showbutton={+isPosterUrlCorrect}
                  color='primary'
                  variant='contained'
                  size='small'
                >
                  {posterSearchLanguage === 'en' ? 'EN' : 'RU'}
                </PosterLanguageSwitch>
              )}

              <ClearPosterButton
                showbutton={+isPosterUrlCorrect}
                onClick={() => {
                  removePoster()
                  setIsUserInteractedWithPoster(true)
                }}
                color='primary'
                variant='contained'
                size='small'
              >
                {t('Clear')}
              </ClearPosterButton>
            </PosterWrapper>
          </RightSideContainer>

          <RightSideContainer
            isError={torrentSource && !isTorrentSourceCorrect}
            notificationMessage={
              !torrentSource ? t('AddTorrentSourceNotification') : !isTorrentSourceCorrect && t('WrongTorrentSource')
            }
            // isHidden={isTorrentSourceCorrect}
            isHidden
          />
        </RightSide>
      </Content>

      <ButtonWrapper>
        <Button onClick={handleClose} color='primary' variant='outlined'>
          {t('Cancel')}
        </Button>

        <Button
          variant='contained'
          style={{ minWidth: '110px' }}
          disabled={!torrentSource}
          onClick={handleSave}
          color='primary'
        >
          {isLoadingButton ? <CircularProgress style={{ color: 'white' }} size={20} /> : t('Add')}
        </Button>
      </ButtonWrapper>
    </Dialog>
  )
}
