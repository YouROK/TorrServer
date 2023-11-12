import { useTranslation } from 'react-i18next'
import { rgba } from 'polished'
import { NoImageIcon } from 'icons'
import { IconButton, InputAdornment, TextField, useTheme } from '@material-ui/core'
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
import { checkImageURL } from './helpers'

export default function RightSideComponent({
  setTitle,
  setPosterUrl,
  setIsPosterUrlCorrect,
  setIsUserInteractedWithPoster,
  setPosterList,
  isTorrentSourceCorrect,
  isHashAlreadyExists,
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
  isEditMode,
}) {
  const { t } = useTranslation()
  const primary = useTheme().palette.primary.main

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
      <RightSideContainer isHidden={!isTorrentSourceCorrect || (isHashAlreadyExists && !isEditMode)}>
        {originalTorrentTitle ? (
          <>
            <TextField
              value={originalTorrentTitle}
              margin='dense'
              label={t('AddDialog.OriginalTorrentTitle')}
              type='text'
              variant='outlined'
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
              label={t('AddDialog.CustomTorrentTitle')}
              type='text'
              variant='outlined'
              fullWidth
              helperText={t('AddDialog.CustomTorrentTitleHelperText')}
              InputProps={{
                endAdornment: (
                  <InputAdornment position='end'>
                    <IconButton
                      style={{ padding: '1px' }}
                      onClick={() => {
                        setTitle('')
                        setIsCustomTitleEnabled(!isCustomTitleEnabled)
                        updateTitleFromSource()
                        setIsUserInteractedWithPoster(false)
                      }}
                    >
                      <HighlightOffIcon style={{ color: isCustomTitleEnabled ? primary : rgba('#ccc', 0.5) }} />
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
          </>
        ) : (
          <TextField
            onChange={handleTitleChange}
            value={title}
            margin='dense'
            label={t('AddDialog.TitleBlank')}
            type='text'
            variant='outlined'
            fullWidth
            helperText={t('AddDialog.TitleBlankHelperText')}
          />
        )}
        <TextField
          onChange={handlePosterUrlChange}
          value={posterUrl}
          margin='dense'
          label={t('AddDialog.AddPosterLinkInput')}
          type='url'
          variant='outlined'
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
                posterSearch(isCustomTitleEnabled ? title : originalTorrentTitle ? parsedTitle : title, newLanguage, {
                  shouldRefreshMainPoster: true,
                })
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
        isError={torrentSource && (!isTorrentSourceCorrect || isHashAlreadyExists)}
        notificationMessage={
          !torrentSource
            ? t('AddDialog.AddTorrentSourceNotification')
            : !isTorrentSourceCorrect
            ? t('AddDialog.WrongTorrentSource')
            : isHashAlreadyExists && t('AddDialog.HashExists')
        }
        isHidden={isEditMode || (isTorrentSourceCorrect && !isHashAlreadyExists)}
      />
    </RightSide>
  )
}
