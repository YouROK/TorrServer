import { useCallback, useMemo, useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogTitle from '@material-ui/core/DialogTitle'
import { torrentsHost, torrentUploadHost } from 'utils/Hosts'
import axios from 'axios'
import { useTranslation } from 'react-i18next'
import { Input, ListItem, ListItemIcon, ListItemText } from '@material-ui/core'
import styled, { css } from 'styled-components'
import { NoImageIcon, AddItemIcon, TorrentIcon } from 'icons'
import debounce from 'lodash/debounce'
import { v4 as uuidv4 } from 'uuid'
import useChangeLanguage from 'utils/useChangeLanguage'
import { Cancel as CancelIcon } from '@material-ui/icons'

const AddDialogStyle = styled.div``
const TitleSection = styled.div`
  background: #00a572;
  color: rgba(0, 0, 0, 0.87);
  font-size: 20px;
  color: #fff;
  font-weight: 500;
  box-shadow: 0px 2px 4px -1px rgb(0 0 0 / 20%), 0px 4px 5px 0px rgb(0 0 0 / 14%), 0px 1px 10px 0px rgb(0 0 0 / 12%);
  padding: 15px 24px;
  position: relative;
`
const MainSection = styled.div`
  background: linear-gradient(145deg, #e4f6ed, #b5dec9);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
`

const MainSectionContentWrapper = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border-bottom: 1px solid rgba(0, 0, 0, 0.12);
`

const LeftSide = styled.div`
  padding: 0 20px 20px 20px;
  border-right: 1px solid rgba(0, 0, 0, 0.12);
`
const RightSide = styled.div`
  display: flex;
  flex-direction: column;
`

const RightSideBottomSectionBasicStyles = css`
  transition: all 0.3s;
  padding: 20px;
  height: 100%;
  display: grid;
`

const RightSideBottomSectionNoFile = styled.div`
  ${RightSideBottomSectionBasicStyles}

  justify-items: center;
  grid-template-rows: 100px 1fr;
  cursor: pointer;

  :hover {
    background-color: rgba(0, 0, 0, 0.04);
    svg {
      transform: translateY(-4%);
    }
  }
`

const RightSideBottomSectionFileSelected = styled.div`
  ${RightSideBottomSectionBasicStyles}
  place-items: center;
`

const TorrentIconWrapper = styled.div`
  position: relative;
`

const CancelIconWrapper = styled.div`
  position: absolute;
  top: -9px;
  left: 10px;
  cursor: pointer;

  > svg {
    transition: all 0.3s;
    fill: rgba(0, 0, 0, 0.7);

    :hover {
      fill: rgba(0, 0, 0, 0.6);
    }
  }
`

const IconWrapper = styled.div`
  display: grid;
  justify-items: center;
  align-content: start;
  gap: 10px;

  svg {
    transition: all 0.3s;
  }
`

const FileUploadLabel = styled.label`
  transition: all 0.3s;
`

const RightSideTopSection = styled.div`
  background: #fff;
  padding: 0 20px 20px 20px;

  ${({ active }) =>
    active &&
    css`
      + ${FileUploadLabel} {
        box-shadow: inset 3px 25px 8px -25px rgba(0, 0, 0, 0.5);
      }
    `};
`

const PosterWrapper = styled.div`
  margin-top: 20px;
  display: grid;
  grid-template-columns: max-content 1fr;
  grid-template-rows: 300px max-content;
  gap: 5px;
`
const PosterSuggestions = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 70px);
  grid-template-rows: repeat(4, calc(25% - 4px));
  grid-auto-flow: column;
  gap: 5px;
`
const PosterSuggestionsItem = styled.div`
  cursor: pointer;

  img {
    transition: all 0.3s;
    border-radius: 5px;
    width: 100%;
    height: 100%;
    object-fit: cover;

    :hover {
      filter: brightness(130%);
    }
  }
`

export const Poster = styled.div`
  ${({ poster }) => css`
    border-radius: 5px;
    overflow: hidden;
    width: 200px;

    ${poster
      ? css`
          img {
            width: 200px;
            object-fit: cover;
            border-radius: 5px;
            height: 100%;
          }
        `
      : css`
          display: grid;
          place-items: center;
          background: #74c39c;

          svg {
            transform: scale(1.5) translateY(-3px);
          }
        `}

    ${
      '' /* @media (max-width: 1280px) {
      align-self: start;
    }

    @media (max-width: 840px) {
      ${poster
        ? css`
            height: 200px;
          `
        : css`
            display: none;
          `}
    } */
    }
  `}
`

const ButtonWrapper = styled.div`
  padding: 20px;
  display: flex;
  justify-content: flex-end;

  > :not(:last-child) {
    margin-right: 10px;
  }
`

const getMoviePosters = (movieName, language = 'en') => {
  const request = `${`http://api.themoviedb.org/3/search/multi?api_key=${process.env.REACT_APP_TMDB_API_KEY}`}&language=${language}&include_image_language=${language},null&query=${movieName}`

  return axios
    .get(request)
    .then(({ data: { results } }) =>
      results.filter(el => el.poster_path).map(el => `https://image.tmdb.org/t/p/w300${el.poster_path}`),
    )
    .catch(() => null)
}

const checkImageURL = async url => {
  if (!url || !url.match(/.(jpg|jpeg|png|gif)$/i)) return false

  try {
    await fetch(url)
    return true
  } catch (e) {
    return false
  }
}

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
      const data = new FormData()
      data.append('save', 'true')
      data.append('file', selectedFile)
      posterUrl && data.append('poster', posterUrl)
      axios.post(torrentUploadHost(), data).finally(() => handleClose())
    } else {
      axios
        .post(torrentsHost(), { action: 'add', link: torrentSource, title, poster: posterUrl, save_to_db: true })
        .finally(() => handleClose())
    }
  }

  const handleCapture = ({ target: { files } }) => {
    const [file] = files
    if (!file) return

    setSelectedFile(file)
    setTorrentSource(file.name)
  }

  const clearSelectedFile = e => {
    e.stopPropagation()
    setSelectedFile()
    setTorrentSource('')
  }

  const userChangesPosterUrl = url => {
    setPosterUrl(url)
    checkImageURL(url).then(setIsPosterUrlCorrect)
    setIsUserInteractedWithPoster(true)
  }

  return (
    <>
      <Dialog open onClose={handleClose} aria-labelledby='form-dialog-title' fullWidth maxWidth='md'>
        <AddDialogStyle>
          <TitleSection>Add new torrent</TitleSection>
          <MainSection>
            <MainSectionContentWrapper>
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

                  <Button
                    style={{ justifySelf: 'center' }}
                    onClick={removePoster}
                    color='primary'
                    variant='outlined'
                    size='small'
                    disabled={!posterUrl}
                  >
                    Clear
                  </Button>
                </PosterWrapper>
              </LeftSide>

              <RightSide>
                <RightSideTopSection active={torrentSourceSelected}>
                  <TextField
                    onChange={handleTorrentSourceChange}
                    value={torrentSource}
                    margin='dense'
                    label={t('TorrentSourceLink')}
                    helperText='magnet / hash / .torrent file link'
                    type='text'
                    fullWidth
                    onFocus={() => setTorrentSourceSelected(true)}
                    onBlur={() => setTorrentSourceSelected(false)}
                    inputProps={{ autoComplete: 'off' }}
                    disabled={selectedFile}
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
                  <FileUploadLabel htmlFor='upload-file' style={{ flex: 1 }}>
                    <input
                      onChange={handleCapture}
                      accept='.torrent'
                      type='file'
                      style={{ display: 'none' }}
                      id='upload-file'
                    />

                    <RightSideBottomSectionNoFile selectedFile={selectedFile} type='submit'>
                      <div>OR</div>

                      <IconWrapper>
                        <AddItemIcon color='primary' />
                        <div>CLICK / DRAG & DROP</div>
                      </IconWrapper>
                    </RightSideBottomSectionNoFile>
                  </FileUploadLabel>
                )}
              </RightSide>
            </MainSectionContentWrapper>
            <ButtonWrapper>
              <Button onClick={handleClose} color='primary' variant='outlined'>
                {t('Cancel')}
              </Button>

              <Button variant='contained' disabled={!torrentSource} onClick={handleSave} color='primary'>
                {t('Add')}
              </Button>
            </ButtonWrapper>
          </MainSection>
        </AddDialogStyle>
        {/* <DialogTitle id='form-dialog-title'>{t('AddMagnetOrLink')}</DialogTitle>
  
        <DialogContent>
          <TextField onChange={inputTitle} margin='dense' id='title' label={t('Title')} type='text' fullWidth />
          <TextField onChange={inputPoster} margin='dense' id='poster' label={t('Poster')} type='url' fullWidth />
          <TextField
            onChange={inputMagnet}
            autoFocus
            margin='dense'
            id='magnet'
            label={t('MagnetOrTorrentFileLink')}
            type='text'
            fullWidth
          />
  
          <Button color='primary' variant='outlined' component='label'>
            {t('UploadFile')}
            <input onChange={handleCapture} type='file' accept='.torrent' hidden />
          </Button> */}
        {/* <label htmlFor='upload-file'>
          <Input onChange={handleCapture} accept='.torrent' type='file' id='upload-file' />
          <Button htmlFor='upload-file' type='submit' color='primary' variant='outlined'>
            {t('UploadFile')}
          </Button>
          <ListItem button variant='raised' type='submit' component='span' key={t('UploadFile')}>
            <ListItemIcon>
              <PublishIcon />
            </ListItemIcon>

            <ListItemText primary={t('UploadFile')} />
          </ListItem>
        </label> */}
        {/* </DialogContent>
  
        <DialogActions>
          <Button onClick={handleClose} color='primary' variant='outlined'>
            {t('Cancel')}
          </Button>
  
          <Button variant='contained' disabled={!link} onClick={handleSave} color='primary'>
            {t('Add')}
          </Button>
        </DialogActions> */}
      </Dialog>
    </>
  )
}
