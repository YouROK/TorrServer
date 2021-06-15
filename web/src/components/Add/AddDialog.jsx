import { useEffect, useMemo, useState } from 'react'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import { torrentsHost, torrentUploadHost } from 'utils/Hosts'
import axios from 'axios'
import { useTranslation } from 'react-i18next'
import debounce from 'lodash/debounce'
import useChangeLanguage from 'utils/useChangeLanguage'
import { useMediaQuery } from '@material-ui/core'
import CircularProgress from '@material-ui/core/CircularProgress'
import usePreviousState from 'utils/usePreviousState'

import { checkImageURL, getMoviePosters, chechTorrentSource, parseTorrentTitle } from './helpers'
import { ButtonWrapper, Content, Header } from './style'
import RightSideComponent from './RightSideComponent'
import LeftSideComponent from './LeftSideComponent'

export default function AddDialog({
  handleClose,
  hash: originalHash,
  title: originalTitle,
  name: originalName,
  poster: originalPoster,
}) {
  const { t } = useTranslation()
  const [torrentSource, setTorrentSource] = useState(originalHash || '')
  const [title, setTitle] = useState(originalTitle || '')
  const [parsedTitle, setParsedTitle] = useState('')
  const [posterUrl, setPosterUrl] = useState(originalPoster || '')
  const [isPosterUrlCorrect, setIsPosterUrlCorrect] = useState(false)
  const [isTorrentSourceCorrect, setIsTorrentSourceCorrect] = useState(false)
  const [posterList, setPosterList] = useState()
  const [isUserInteractedWithPoster, setIsUserInteractedWithPoster] = useState(false)
  const [currentLang] = useChangeLanguage()
  const [selectedFile, setSelectedFile] = useState()
  const [posterSearchLanguage, setPosterSearchLanguage] = useState(currentLang === 'ru' ? 'ru' : 'en')
  const [isLoadingButton, setIsLoadingButton] = useState(false)
  const [skipDebounce, setSkipDebounce] = useState(false)
  const [isEditMode, setIsEditMode] = useState(false)

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  useEffect(() => {
    if (originalHash) {
      setIsEditMode(true)

      checkImageURL(posterUrl).then(correctImage => {
        correctImage ? setIsPosterUrlCorrect(true) : removePoster()
      })
    }
    // This is needed only on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const posterSearch = useMemo(
    () =>
      (movieName, language, { shouldRefreshMainPoster = false } = {}) => {
        if (!movieName) {
          setPosterList()
          removePoster()
          return
        }

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

  const delayedPosterSearch = useMemo(() => debounce(posterSearch, 700), [posterSearch])

  const prevParsedTitleState = usePreviousState(parsedTitle)
  const prevTorrentSourceState = usePreviousState(torrentSource)

  useEffect(() => {
    // if torrentSource is updated then we are checking that source is valid and getting title from the source
    const torrentSourceChanged = torrentSource !== prevTorrentSourceState

    const isCorrectSource = chechTorrentSource(torrentSource)
    if (!isCorrectSource) return setIsTorrentSourceCorrect(false)

    setIsTorrentSourceCorrect(true)

    if (torrentSourceChanged) {
      parseTorrentTitle(selectedFile || torrentSource, ({ parsedTitle, originalName }) => {
        if (!parsedTitle) return

        setSkipDebounce(true)
        setTitle(originalName)
        setParsedTitle(parsedTitle)
      })
    }
  }, [prevTorrentSourceState, selectedFile, torrentSource])

  useEffect(() => {
    // if title exists and title was changed then search poster.
    const titleChanged = parsedTitle !== prevParsedTitleState
    if (!titleChanged) return

    if (skipDebounce) {
      posterSearch(parsedTitle, posterSearchLanguage)
      setSkipDebounce(false)
    } else {
      parsedTitle === '' ? removePoster() : delayedPosterSearch(parsedTitle, posterSearchLanguage)
    }
  }, [parsedTitle, prevParsedTitleState, delayedPosterSearch, posterSearch, posterSearchLanguage, skipDebounce])

  const removePoster = () => {
    setIsPosterUrlCorrect(false)
    setPosterUrl('')
  }

  useEffect(() => {
    if (!selectedFile && !torrentSource) {
      setTitle('')
      setParsedTitle('')
      setPosterList()
      removePoster()
      setIsUserInteractedWithPoster(false)
    }
  }, [selectedFile, torrentSource])

  const handleSave = () => {
    setIsLoadingButton(true)

    if (isEditMode) {
      axios
        .post(torrentsHost(), {
          action: 'set',
          hash: originalHash,
          title: title === '' ? originalName : title,
          poster: posterUrl,
        })
        .finally(handleClose)
    } else if (selectedFile) {
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

  return (
    <Dialog
      open
      onClose={handleClose}
      aria-labelledby='form-dialog-title'
      fullScreen={fullScreen}
      fullWidth
      maxWidth='md'
    >
      <Header>{t(isEditMode ? 'EditTorrent' : 'AddNewTorrent')}</Header>

      <Content isEditMode={isEditMode}>
        {!isEditMode && (
          <LeftSideComponent
            setIsUserInteractedWithPoster={setIsUserInteractedWithPoster}
            setSelectedFile={setSelectedFile}
            torrentSource={torrentSource}
            setTorrentSource={setTorrentSource}
            selectedFile={selectedFile}
          />
        )}

        <RightSideComponent
          setTitle={setTitle}
          setParsedTitle={setParsedTitle}
          setPosterUrl={setPosterUrl}
          setIsPosterUrlCorrect={setIsPosterUrlCorrect}
          setIsUserInteractedWithPoster={setIsUserInteractedWithPoster}
          setPosterList={setPosterList}
          isTorrentSourceCorrect={isTorrentSourceCorrect}
          title={title}
          parsedTitle={parsedTitle}
          posterUrl={posterUrl}
          isPosterUrlCorrect={isPosterUrlCorrect}
          posterList={posterList}
          currentLang={currentLang}
          posterSearchLanguage={posterSearchLanguage}
          setPosterSearchLanguage={setPosterSearchLanguage}
          posterSearch={posterSearch}
          removePoster={removePoster}
          torrentSource={torrentSource}
        />
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
          {isLoadingButton ? <CircularProgress style={{ color: 'white' }} size={20} /> : t(isEditMode ? 'Save' : 'Add')}
        </Button>
      </ButtonWrapper>
    </Dialog>
  )
}
