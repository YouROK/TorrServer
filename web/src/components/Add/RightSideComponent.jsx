import { useTranslation } from 'react-i18next'
import { NoImageIcon } from 'icons'
import { IconButton, InputAdornment, TextField } from '@material-ui/core'
import { CheckBox as CheckBoxIcon } from '@material-ui/icons'
import { useState } from 'react'

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
import { checkImageURL } from './helpers'

export default function RightSideComponent({
  setTitle,
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
  originalTorrentTitle,
  updateTitleFromSource,
  isCustomTitleEnabled,
  setIsCustomTitleEnabled,
}) {
  const { t } = useTranslation()

  const handleTitleChange = ({ target: { value } }) => setTitle(value)
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

  return (
    <RightSide>
      <RightSideContainer isHidden={!isTorrentSourceCorrect}>
        <TextField
          value={originalTorrentTitle}
          margin='dense'
          // label={t('Title')}
          label='Оригинальное название торрента'
          type='text'
          fullWidth
          disabled={isCustomTitleEnabled}
          InputProps={{ readOnly: true }}
        />
        <TextField
          onChange={handleTitleChange}
          onFocus={() => setIsCustomTitleEnabled(true)}
          onBlur={({ target: { value } }) => !value && setIsCustomTitleEnabled(false)}
          value={title}
          margin='dense'
          label='Использовать свое название (не обязательно)'
          type='text'
          fullWidth
          InputProps={{
            endAdornment: (
              <InputAdornment position='end'>
                <IconButton
                  style={{ padding: '0 0 0 7px' }}
                  onClick={() => {
                    setTitle('')
                    setIsCustomTitleEnabled(!isCustomTitleEnabled)
                    updateTitleFromSource()
                    setIsUserInteractedWithPoster(false)
                  }}
                >
                  <CheckBoxIcon style={{ color: isCustomTitleEnabled ? 'green' : 'gray' }} />
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
                posterSearch(isCustomTitleEnabled ? title : parsedTitle, newLanguage, { shouldRefreshMainPoster: true })
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
