import { useCallback, useEffect, useMemo, useState } from 'react'
import Button from '@material-ui/core/Button'
import { torrentsHost, torrentUploadHost } from 'utils/Hosts'
import axios from 'axios'
import { useTranslation } from 'react-i18next'
import debounce from 'lodash/debounce'
import useChangeLanguage from 'utils/useChangeLanguage'
import { useMediaQuery } from '@material-ui/core'
import CircularProgress from '@material-ui/core/CircularProgress'
import usePreviousState from 'utils/usePreviousState'
import { useQuery } from 'react-query'
import { getTorrents } from 'utils/Utils'
import parseTorrent from 'parse-torrent'
import { ButtonWrapper } from 'style/DialogStyles'
import { StyledDialog, StyledHeader } from 'style/CustomMaterialUiStyles'
import useOnStandaloneAppOutsideClick from 'utils/useOnStandaloneAppOutsideClick'

import { checkImageURL, getMoviePosters, checkTorrentSource, parseTorrentTitle } from './helpers'
import { Content } from './style'
import RightSideComponent from './RightSideComponent'
import LeftSideComponent from './LeftSideComponent'

export default function AddDialog({
  handleClose,
  hash: originalHash,
  title: originalTitle,
  name: originalName,
  poster: originalPoster,
  category: originalCategory,
}) {
  const { t } = useTranslation()
  const isEditMode = !!originalHash
  const [torrentSource, setTorrentSource] = useState(originalHash || '')
  const [title, setTitle] = useState(originalTitle || '')
  const [category, setCategory] = useState(originalCategory || '')
  const [originalTorrentTitle, setOriginalTorrentTitle] = useState('')
  const [parsedTitle, setParsedTitle] = useState('')
  const [posterUrl, setPosterUrl] = useState(originalPoster || '')
  const [isPosterUrlCorrect, setIsPosterUrlCorrect] = useState(false)
  const [isTorrentSourceCorrect, setIsTorrentSourceCorrect] = useState(false)
  const [isHashAlreadyExists, setIsHashAlreadyExists] = useState(false)
  const [posterList, setPosterList] = useState()
  const [isUserInteractedWithPoster, setIsUserInteractedWithPoster] = useState(isEditMode)
  const [currentLang] = useChangeLanguage()
  const [selectedFile, setSelectedFile] = useState()
  const [posterSearchLanguage, setPosterSearchLanguage] = useState(currentLang === 'ru' ? 'ru' : 'en')
  const [isSaving, setIsSaving] = useState(false)
  const [skipDebounce, setSkipDebounce] = useState(false)
  const [isCustomTitleEnabled, setIsCustomTitleEnabled] = useState(false)
  const [currentSourceHash, setCurrentSourceHash] = useState()

  const ref = useOnStandaloneAppOutsideClick(handleClose)

  const { data: torrents } = useQuery('torrents', getTorrents, { retry: 1, refetchInterval: 1000 })

  useEffect(() => {
    // getting hash from added torrent source
    parseTorrent.remote(selectedFile || torrentSource, (_, { infoHash } = {}) => setCurrentSourceHash(infoHash))
  }, [selectedFile, torrentSource])

  useEffect(() => {
    // checking if torrent already exists in DB
    if (!setCurrentSourceHash) return

    const allHashes = torrents.map(({ hash }) => hash)
    setIsHashAlreadyExists(allHashes.includes(currentSourceHash))
  }, [currentSourceHash, torrents])

  useEffect(() => {
    // closing dialog when torrent successfully added in DB
    if (!isSaving) return

    const allHashes = torrents.map(({ hash }) => hash)
    allHashes.includes(currentSourceHash) && handleClose()
    // FIXME! check api reply on add links
    const linkRegex = /^(http(s?)):\/\/.*/i
    torrentSource.match(linkRegex) !== null && handleClose()
  }, [isSaving, torrents, torrentSource, currentSourceHash, handleClose])

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  const updateTitleFromSource = useCallback(() => {
    parseTorrentTitle(selectedFile || torrentSource, ({ parsedTitle, originalName }) => {
      if (!originalName) return

      setSkipDebounce(true)
      setTitle('')
      setIsCustomTitleEnabled(false)
      setOriginalTorrentTitle(originalName)
      setParsedTitle(parsedTitle)
    })
  }, [selectedFile, torrentSource])

  useEffect(() => {
    if (!selectedFile && !torrentSource) {
      setTitle('')
      setOriginalTorrentTitle('')
      setParsedTitle('')
      setIsCustomTitleEnabled(false)
      setPosterList()
      removePoster()
      setIsUserInteractedWithPoster(false)
    }
  }, [selectedFile, torrentSource])

  const removePoster = () => {
    setIsPosterUrlCorrect(false)
    setPosterUrl('')
  }

  useEffect(() => {
    if (originalHash) {
      checkImageURL(posterUrl).then(correctImage => {
        correctImage ? setIsPosterUrlCorrect(true) : removePoster()
      })
    }
    // This is needed only on mount. Do not remove line below
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

  const prevTorrentSourceState = usePreviousState(torrentSource)

  useEffect(() => {
    const isCorrectSource = checkTorrentSource(torrentSource)
    if (!isCorrectSource) return setIsTorrentSourceCorrect(false)

    setIsTorrentSourceCorrect(true)

    // if torrentSource is updated then we are getting title from the source
    const torrentSourceChanged = torrentSource !== prevTorrentSourceState
    if (!torrentSourceChanged) return

    updateTitleFromSource()
  }, [prevTorrentSourceState, selectedFile, torrentSource, updateTitleFromSource])

  const prevTitleState = usePreviousState(title)

  useEffect(() => {
    // if title exists and title was changed then search poster.
    const titleChanged = title !== prevTitleState
    if (!titleChanged && !parsedTitle) return

    if (skipDebounce) {
      posterSearch(title || parsedTitle, posterSearchLanguage)
      setSkipDebounce(false)
    } else if (!title) {
      delayedPosterSearch.cancel()

      if (parsedTitle) {
        posterSearch(parsedTitle, posterSearchLanguage)
      } else {
        !isUserInteractedWithPoster && removePoster()
      }
    } else {
      delayedPosterSearch(title, posterSearchLanguage)
    }
  }, [
    title,
    parsedTitle,
    prevTitleState,
    delayedPosterSearch,
    posterSearch,
    posterSearchLanguage,
    skipDebounce,
    isUserInteractedWithPoster,
  ])

  const handleSave = () => {
    setIsSaving(true)

    if (isEditMode) {
      axios
        .post(torrentsHost(), {
          action: 'set',
          hash: originalHash,
          title: title || originalName,
          poster: posterUrl,
          category,
        })
        .finally(handleClose)
    } else if (selectedFile) {
      // file save
      const data = new FormData()
      data.append('save', 'true')
      data.append('file', selectedFile)
      title && data.append('title', title)
      category && data.append('category', category)
      posterUrl && data.append('poster', posterUrl)
      axios.post(torrentUploadHost(), data).catch(handleClose)
    } else {
      // link save
      axios
        .post(torrentsHost(), {
          action: 'add',
          link: torrentSource,
          title,
          category,
          poster: posterUrl,
          save_to_db: true,
        })
        .catch(handleClose)
    }
  }

  return (
    <StyledDialog open onClose={handleClose} fullScreen={fullScreen} fullWidth maxWidth='md' ref={ref}>
      <StyledHeader>{t(isEditMode ? 'EditTorrent' : 'AddNewTorrent')}</StyledHeader>

      <Content isEditMode={isEditMode}>
        {!isEditMode && (
          <LeftSideComponent
            setIsUserInteractedWithPoster={setIsUserInteractedWithPoster}
            selectedFile={selectedFile}
            setSelectedFile={setSelectedFile}
            torrentSource={torrentSource}
            setTorrentSource={setTorrentSource}
          />
        )}
        <RightSideComponent
          originalTorrentTitle={originalTorrentTitle}
          setTitle={setTitle}
          setCategory={setCategory}
          setPosterUrl={setPosterUrl}
          setIsPosterUrlCorrect={setIsPosterUrlCorrect}
          setIsUserInteractedWithPoster={setIsUserInteractedWithPoster}
          setPosterList={setPosterList}
          isTorrentSourceCorrect={isTorrentSourceCorrect}
          isHashAlreadyExists={isHashAlreadyExists}
          title={title}
          category={category}
          parsedTitle={parsedTitle}
          posterUrl={posterUrl}
          isPosterUrlCorrect={isPosterUrlCorrect}
          posterList={posterList}
          currentLang={currentLang}
          posterSearchLanguage={posterSearchLanguage}
          setPosterSearchLanguage={setPosterSearchLanguage}
          posterSearch={posterSearch}
          removePoster={removePoster}
          updateTitleFromSource={updateTitleFromSource}
          torrentSource={torrentSource}
          isCustomTitleEnabled={isCustomTitleEnabled}
          setIsCustomTitleEnabled={setIsCustomTitleEnabled}
          isEditMode={isEditMode}
        />
      </Content>

      <ButtonWrapper>
        <Button onClick={handleClose} color='secondary' variant='outlined'>
          {t('Cancel')}
        </Button>

        <Button
          variant='contained'
          style={{ minWidth: '110px' }}
          disabled={!torrentSource || (isHashAlreadyExists && !isEditMode) || !isTorrentSourceCorrect}
          onClick={handleSave}
          color='secondary'
        >
          {isSaving ? <CircularProgress style={{ color: 'white' }} size={20} /> : t(isEditMode ? 'Save' : 'Add')}
        </Button>
      </ButtonWrapper>
    </StyledDialog>
  )
}
