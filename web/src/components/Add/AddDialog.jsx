import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import Button from '@material-ui/core/Button'
import { torrentsHost } from 'utils/Hosts'
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
import ptt from 'parse-torrent-title'
import { ButtonWrapper } from 'style/DialogStyles'
import { StyledDialog, StyledHeader } from 'style/CustomMaterialUiStyles'
import useOnStandaloneAppOutsideClick from 'utils/useOnStandaloneAppOutsideClick'

import {
  checkImageURL,
  getMoviePosters,
  checkTorrentSource,
  parseTorrentTitle,
  shortenTitleForPosterSearch,
} from './helpers'
import { Content } from './style'
import RightSideComponent from './RightSideComponent'
import LeftSideComponent from './LeftSideComponent'
import MultiAddDialog from './MultiAddDialog'

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
  const [posterSearchLanguage, setPosterSearchLanguage] = useState(currentLang === 'ru' ? 'ru' : 'en')
  const [isSaving, setIsSaving] = useState(false)
  const [skipDebounce, setSkipDebounce] = useState(false)
  const [isCustomTitleEnabled, setIsCustomTitleEnabled] = useState(false)
  const [currentSourceHash, setCurrentSourceHash] = useState()
  const editModePosterSearchedRef = useRef(false)

  // When files are dropped/selected, switch to MultiAddDialog
  const [multiFiles, setMultiFiles] = useState(null)

  const ref = useOnStandaloneAppOutsideClick(handleClose)

  const { data: torrents } = useQuery('torrents', getTorrents, { retry: 1, refetchInterval: 1000 })

  useEffect(() => {
    parseTorrent.remote(torrentSource, (_, { infoHash } = {}) => setCurrentSourceHash(infoHash))
  }, [torrentSource])

  useEffect(() => {
    if (!currentSourceHash || !torrents) return

    const allHashes = torrents.map(({ hash }) => hash)
    setIsHashAlreadyExists(allHashes.includes(currentSourceHash))
  }, [currentSourceHash, torrents])

  useEffect(() => {
    if (!isSaving || !torrents) return

    const allHashes = torrents.map(({ hash }) => hash)
    allHashes.includes(currentSourceHash) && handleClose()
    const linkRegex = /^(http(s?)):\/\/.*/i
    torrentSource.match(linkRegex) !== null && handleClose()
  }, [isSaving, torrents, torrentSource, currentSourceHash, handleClose])

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  const updateTitleFromSource = useCallback(() => {
    parseTorrentTitle(torrentSource, ({ parsedTitle, originalName }) => {
      if (!originalName) return

      setSkipDebounce(true)
      setTitle('')
      setIsCustomTitleEnabled(false)
      setOriginalTorrentTitle(originalName)
      setParsedTitle(parsedTitle)
    })
  }, [torrentSource])

  useEffect(() => {
    if (!torrentSource) {
      setTitle('')
      setOriginalTorrentTitle('')
      setParsedTitle('')
      setIsCustomTitleEnabled(false)
      setPosterList()
      removePoster()
      setIsUserInteractedWithPoster(false)
    }
  }, [torrentSource])

  const removePoster = () => {
    setIsPosterUrlCorrect(false)
    setPosterUrl('')
  }

  // Edit mode: init original/parsed title from name so poster can be searched
  useEffect(() => {
    if (!originalHash || (!originalName && !originalTitle)) return
    const source = originalName || originalTitle
    setOriginalTorrentTitle(source)
    try {
      const parsed = ptt.parse(source)
      setParsedTitle(parsed?.title || '')
    } catch (_) {
      setParsedTitle('')
    }
    editModePosterSearchedRef.current = false
  }, [originalHash, originalName, originalTitle])

  useEffect(() => {
    if (originalHash) {
      checkImageURL(posterUrl).then(correctImage => {
        correctImage ? setIsPosterUrlCorrect(true) : removePoster()
      })
    }
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
        const query = shortenTitleForPosterSearch(String(movieName).trim())

        getMoviePosters(query || movieName, language).then(urlList => {
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

    const torrentSourceChanged = torrentSource !== prevTorrentSourceState
    if (!torrentSourceChanged) return

    updateTitleFromSource()
  }, [prevTorrentSourceState, torrentSource, updateTitleFromSource])

  // Edit mode: auto-search poster once when we have title and no poster
  useEffect(() => {
    if (
      !originalHash ||
      editModePosterSearchedRef.current ||
      originalPoster ||
      !(parsedTitle || originalTitle || title)
    ) {
      return
    }
    const searchTitle = parsedTitle || title || originalTitle
    if (!shortenTitleForPosterSearch(searchTitle)) return
    editModePosterSearchedRef.current = true
    posterSearch(searchTitle, posterSearchLanguage, { shouldRefreshMainPoster: true })
  }, [originalHash, originalPoster, parsedTitle, originalTitle, title, posterSearchLanguage, posterSearch])

  const prevTitleState = usePreviousState(title)

  useEffect(() => {
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

  const handleSetSelectedFile = useCallback(fileOrFiles => {
    const files = Array.isArray(fileOrFiles) ? fileOrFiles : [fileOrFiles]
    setMultiFiles(files)
  }, [])

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

  if (multiFiles) {
    return <MultiAddDialog files={multiFiles} handleClose={handleClose} />
  }

  return (
    <StyledDialog open onClose={handleClose} fullScreen={fullScreen} fullWidth maxWidth='md' ref={ref}>
      <StyledHeader>{t(isEditMode ? 'EditTorrent' : 'AddNewTorrent')}</StyledHeader>

      <Content isEditMode={isEditMode}>
        {!isEditMode && (
          <LeftSideComponent
            setIsUserInteractedWithPoster={setIsUserInteractedWithPoster}
            setSelectedFile={handleSetSelectedFile}
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
