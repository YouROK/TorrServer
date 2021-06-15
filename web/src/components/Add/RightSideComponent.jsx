import { useTranslation } from 'react-i18next'
import { NoImageIcon } from 'icons'
import { IconButton, InputAdornment, TextField } from '@material-ui/core'
import { HighlightOff as HighlightOffIcon } from '@material-ui/icons'

import {
  ClearPosterButton,
  PosterLanguageSwitch,
  RightSide,
  Poster,
  PosterSuggestions,
  PosterSuggestionsItem,
  PosterWrapper,
  RightSideContainer,
} from './style'
import { checkImageURL, hashRegex } from './helpers'

export default function RightSideComponent({
  setTitle,
  setParsedTitle,
  setPosterUrl,
  setIsPosterUrlCorrect,
  setIsUserInteractedWithPoster,
  setPosterList,
  isTorrentSourceCorrect,
  title,
  parsedTitle,
  posterUrl,
  isPosterUrlCorrect,
  posterList,
  currentLang,
  posterSearchLanguage,
  setPosterSearchLanguage,
  posterSearch,
  removePoster,
  torrentSource,
}) {
  const { t } = useTranslation()

  const handleTitleChange = ({ target: { value } }) => {
    setTitle(value)
    setParsedTitle(value)
  }
  const handlePosterUrlChange = ({ target: { value } }) => {
    setPosterUrl(value)
    checkImageURL(value).then(setIsPosterUrlCorrect)
    setIsUserInteractedWithPoster(!!value)
    setPosterList()
  }
  const userChangesPosterUrl = url => {
    setPosterUrl(url)
    checkImageURL(url).then(setIsPosterUrlCorrect)
    setIsUserInteractedWithPoster(true)
  }

  const sourceIsHash = torrentSource.match(hashRegex) !== null

  return (
    <RightSide>
      <RightSideContainer isHidden={!isTorrentSourceCorrect}>
        <TextField
          onChange={handleTitleChange}
          value={title}
          margin='dense'
          label={t(sourceIsHash ? 'AddDialogTorrentTitle' : 'Title')}
          type='text'
          fullWidth
          InputProps={{
            endAdornment:
              title === '' ? null : (
                <InputAdornment position='end'>
                  <IconButton
                    aria-label='clear input'
                    onClick={() => {
                      setTitle('')
                      setParsedTitle('')
                    }}
                  >
                    <HighlightOffIcon />
                  </IconButton>
                </InputAdornment>
              ),
          }}
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
                <PosterSuggestionsItem onClick={() => userChangesPosterUrl(url)} key={url}>
                  <img src={url} alt='poster' />
                </PosterSuggestionsItem>
              ))}
          </PosterSuggestions>

          {currentLang !== 'en' && (
            <PosterLanguageSwitch
              onClick={() => {
                const newLanguage = posterSearchLanguage === 'en' ? 'ru' : 'en'
                setPosterSearchLanguage(newLanguage)
                posterSearch(parsedTitle, newLanguage, { shouldRefreshMainPoster: true })
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
        isHidden={isTorrentSourceCorrect}
      />
    </RightSide>
  )
}
