import axios from 'axios'
import Dialog from '@material-ui/core/Dialog'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import TextField from '@material-ui/core/TextField'
import DialogActions from '@material-ui/core/DialogActions'
import Button from '@material-ui/core/Button'
import Checkbox from '@material-ui/core/Checkbox'
import {
  ButtonGroup,
  FormControlLabel,
  Grid,
  Input,
  InputLabel,
  Select,
  Slider,
  Switch,
  useMediaQuery,
  useTheme,
} from '@material-ui/core'
import { settingsHost, setTorrServerHost, getTorrServerHost } from 'utils/Hosts'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Header } from 'style/DialogStyles'
import AppBar from '@material-ui/core/AppBar'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import SwipeableViews from 'react-swipeable-views'
import styled, { css } from 'styled-components'
import { USBIcon, RAMIcon } from 'icons'

const FooterSection = styled.div`
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #e8e5eb;

  > :last-child > :not(:last-child) {
    margin-right: 10px;
  }
`
const Divider = styled.div`
  height: 1px;
  background-color: rgba(0, 0, 0, 0.12);
  margin: 30px 0;
`

const PreloadCachePercentage = styled.div.attrs(
  ({
    value,
    // theme: {
    //   dialogTorrentDetailsContent: { gradientEndColor },
    // },
  }) => {
    const gradientStartColor = 'lightblue'
    const gradientEndColor = 'orangered'

    return {
      // this block is here according to styled-components recomendation about fast changable components
      style: {
        background: `linear-gradient(to right, ${gradientEndColor} 0%, ${gradientEndColor} ${value}%, ${gradientStartColor} ${value}%, ${gradientStartColor} 100%)`,
      },
    }
  },
)`
  ${({ label, isPreloadEnabled }) => css`
    border: 1px solid;
    padding: 10px 20px;
    border-radius: 5px;
    color: #000;
    margin-bottom: 10px;
    position: relative;

    :before {
      content: '${label}';
      display: grid;
      place-items: center;
      font-size: 20px;
    }

    ${isPreloadEnabled &&
    css`
      :after {
        content: '';
        width: 100%;
        height: 3px;
        background: green;
        position: absolute;
        bottom: 0;
        left: 0;
      }
    `}
  `}
`

const PreloadCacheValue = styled.div`
  ${({ color }) => css`
    display: grid;
    grid-template-columns: max-content 100px 1fr;
    gap: 10px;
    align-items: center;

    :not(:last-child) {
      margin-bottom: 5px;
    }

    :before {
      content: '';
      background: ${color};
      width: 15px;
      height: 15px;
      border-radius: 50%;
    }
  `}
`

const a11yProps = index => ({
  id: `full-width-tab-${index}`,
  'aria-controls': `full-width-tabpanel-${index}`,
})

const TabPanel = ({ children, value, index, ...other }) => (
  <div role='tabpanel' hidden={value !== index} id={`full-width-tabpanel-${index}`} {...other}>
    {value === index && <>{children}</>}
  </div>
)

const MainSettingsContent = styled.div`
  min-height: 500px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 40px;
  padding: 20px;

  @media (max-width: 930px) {
    grid-template-columns: 1fr;
  }
`
const SecondarySettingsContent = styled.div`
  min-height: 500px;
  padding: 20px;
`

const StorageButton = styled.div`
  display: grid;
  place-items: center;
  gap: 10px;
`

const StorageIconWrapper = styled.div`
  ${({ selected }) => css`
    width: 150px;
    height: 150px;
    border-radius: 50%;
    background: ${selected ? 'blue' : 'lightgray'};
    transition: 0.2s;

    ${!selected &&
    css`
      cursor: pointer;

      :hover {
        background: orangered;
      }
    `}

    svg {
      transform: rotate(-45deg) scale(0.75);
    }
  `}
`

const CacheSizeSettings = styled.div``
const CacheStorageSelector = styled.div`
  display: grid;
  grid-template-rows: max-content 1fr;
  grid-template-columns: 1fr 1fr;
  grid-template-areas: 'label label';
  place-items: center;

  @media (max-width: 930px) {
    grid-template-columns: repeat(2, max-content);
    column-gap: 30px;
  }
`

const CacheStorageSettings = styled.div``

const SettingSection = styled.section``
const SettingLabel = styled.div``
const SettingSectionLabel = styled.div`
  font-size: 25px;
  padding-bottom: 20px;
`

export default function SettingsDialog({ handleClose }) {
  const { t } = useTranslation()

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  const [settings, setSets] = useState({})
  const [show, setShow] = useState(false)
  const [tsHost, setTSHost] = useState(getTorrServerHost())
  useEffect(() => {
    axios
      .post(settingsHost(), { action: 'get' })
      .then(({ data }) => {
        setSets({ ...data, CacheSize: data.CacheSize / (1024 * 1024) })
        setShow(true)
      })
      .catch(() => setShow(false))
  }, [tsHost])

  const handleSave = () => {
    handleClose()
    const sets = JSON.parse(JSON.stringify(settings))
    sets.CacheSize *= 1024 * 1024
    axios.post(settingsHost(), { action: 'set', sets })
  }
  const onInputHost = ({ target: { value } }) => {
    const host = value.replace(/\/$/gi, '')
    setTorrServerHost(host)
    setTSHost(host)
  }

  const inputForm = ({ target: { type, value, checked, id } }) => {
    const sets = JSON.parse(JSON.stringify(settings))
    if (type === 'number' || type === 'select-one') {
      sets[id] = Number(value)
    } else if (type === 'checkbox') {
      if (
        id === 'DisableTCP' ||
        id === 'DisableUTP' ||
        id === 'DisableUPNP' ||
        id === 'DisableDHT' ||
        id === 'DisablePEX' ||
        id === 'DisableUpload'
      )
        sets[id] = Boolean(!checked)
      else sets[id] = Boolean(checked)
    } else if (type === 'url') {
      sets[id] = value
    }
    setSets(sets)
  }

  const {
    CacheSize,
    PreloadBuffer,
    ReaderReadAHead,
    RetrackersMode,
    TorrentDisconnectTimeout,
    EnableIPv6,
    ForceEncrypt,
    DisableTCP,
    DisableUTP,
    DisableUPNP,
    DisableDHT,
    DisablePEX,
    DisableUpload,
    DownloadRateLimit,
    UploadRateLimit,
    ConnectionsLimit,
    DhtConnectionLimit,
    PeersListenPort,
    UseDisk,
    TorrentsSavePath,
    RemoveCacheOnDrop,
  } = settings

  const { direction } = useTheme()
  const [selectedTab, setSelectedTab] = useState(0)

  const handleChange = (_, newValue) => setSelectedTab(newValue)

  const handleChangeIndex = index => setSelectedTab(index)

  const [cacheSize, setCacheSize] = useState(96)
  const [cachePercentage, setCachePercentage] = useState(95)
  const [isProMode, setIsProMode] = useState(JSON.parse(localStorage.getItem('isProMode')) || false)
  const [isRamSelected, setIsRamSelected] = useState(true)

  const handleSliderChange = (_, newValue) => {
    setCacheSize(newValue)
  }

  const handleBlur = ({ target: { value } }) => {
    if (value < 32) return setCacheSize(32)
    if (value > 20000) return setCacheSize(20000)

    setCacheSize(Math.round(value / 8) * 8)
  }

  const handleInputChange = ({ target: { value } }) => {
    setCacheSize(value === '' ? '' : Number(value))
  }

  return (
    <Dialog open onClose={handleClose} fullScreen={fullScreen} fullWidth maxWidth='md'>
      <Header>{t('Settings')}</Header>

      <>
        <AppBar position='static' color='default'>
          <Tabs
            value={selectedTab}
            onChange={handleChange}
            indicatorColor='primary'
            textColor='primary'
            variant='fullWidth'
          >
            <Tab label='Основные' {...a11yProps(0)} />

            <Tab
              disabled={!isProMode}
              label={isProMode ? 'Дополнительные' : 'Дополнительные (включите pro mode)'}
              {...a11yProps(1)}
            />
          </Tabs>
        </AppBar>

        <SwipeableViews
          axis={direction === 'rtl' ? 'x-reverse' : 'x'}
          index={selectedTab}
          onChangeIndex={handleChangeIndex}
        >
          <TabPanel value={selectedTab} index={0} dir={direction}>
            <MainSettingsContent>
              <CacheSizeSettings>
                <SettingSectionLabel>Настройки кеша</SettingSectionLabel>

                <PreloadCachePercentage
                  value={100 - cachePercentage}
                  label={`Кеш ${cacheSize} МБ`}
                  isPreloadEnabled={PreloadBuffer}
                />

                <PreloadCacheValue color='orangered'>
                  <div>
                    {100 - cachePercentage}% ({Math.round((cacheSize / 100) * (100 - cachePercentage))} МБ)
                  </div>
                  <div>От кеша будет оставаться позади воспроизводимого блока</div>
                </PreloadCacheValue>

                <PreloadCacheValue color='lightblue'>
                  <div>
                    {cachePercentage}% ({Math.round((cacheSize / 100) * cachePercentage)} МБ)
                  </div>
                  <div>От кеша будет спереди от воспроизводимого блока</div>
                </PreloadCacheValue>

                <Divider />

                <SettingSection>
                  <SettingLabel>Размер кеша</SettingLabel>

                  <Grid container spacing={2} alignItems='center'>
                    <Grid item xs>
                      <Slider min={32} max={1024} value={cacheSize} onChange={handleSliderChange} step={8} />
                    </Grid>

                    {isProMode && (
                      <Grid item>
                        <Input
                          value={cacheSize}
                          margin='dense'
                          onChange={handleInputChange}
                          onBlur={handleBlur}
                          style={{ width: '65px' }}
                          inputProps={{
                            step: 8,
                            min: 32,
                            max: 20000,
                            type: 'number',
                          }}
                        />
                      </Grid>
                    )}
                  </Grid>
                </SettingSection>

                <SettingSection>
                  <SettingLabel>Кеш предзагрузки</SettingLabel>

                  <Grid container spacing={2} alignItems='center'>
                    <Grid item xs>
                      <Slider
                        min={40}
                        max={95}
                        value={cachePercentage}
                        onChange={(_, newValue) => setCachePercentage(newValue)}
                      />
                    </Grid>

                    {isProMode && (
                      <Grid item>
                        <Input
                          value={cachePercentage}
                          margin='dense'
                          onChange={({ target: { value } }) => setCachePercentage(value === '' ? '' : Number(value))}
                          onBlur={({ target: { value } }) => {
                            if (value < 0) return setCachePercentage(0)
                            if (value > 100) return setCachePercentage(100)
                          }}
                          style={{ width: '65px' }}
                          inputProps={{
                            min: 0,
                            max: 100,
                            type: 'number',
                          }}
                        />
                      </Grid>
                    )}
                  </Grid>
                </SettingSection>

                <SettingSection>
                  <FormControlLabel
                    control={<Switch checked={PreloadBuffer} onChange={inputForm} id='PreloadBuffer' color='primary' />}
                    label={t('PreloadBuffer')}
                  />
                </SettingSection>
              </CacheSizeSettings>

              {isRamSelected ? (
                <CacheStorageSelector>
                  <SettingSectionLabel style={{ placeSelf: 'start', gridArea: 'label' }}>
                    Место хранения кеша
                  </SettingSectionLabel>

                  <StorageButton>
                    <StorageIconWrapper selected>
                      <RAMIcon />
                    </StorageIconWrapper>
                    <div>Оперативная память</div>
                  </StorageButton>

                  <StorageButton>
                    <StorageIconWrapper onClick={() => setIsRamSelected(false)}>
                      <USBIcon />
                    </StorageIconWrapper>
                    <div>Диск</div>
                  </StorageButton>
                </CacheStorageSelector>
              ) : (
                <CacheStorageSettings>
                  <SettingSectionLabel>Место хранения кеша</SettingSectionLabel>

                  <ButtonGroup fullWidth color='primary'>
                    <Button onClick={() => setIsRamSelected(true)}>
                      <div>
                        <RAMIcon width='50px' />
                        <div>Оперативная память</div>
                      </div>
                    </Button>

                    <Button variant='contained'>
                      <div>
                        <USBIcon width='50px' color='white' />
                        <div>Диск</div>
                      </div>
                    </Button>
                  </ButtonGroup>

                  <FormControlLabel
                    control={
                      <Switch checked={RemoveCacheOnDrop} onChange={inputForm} id='RemoveCacheOnDrop' color='primary' />
                    }
                    label={t('RemoveCacheOnDrop')}
                  />
                  <small>{t('RemoveCacheOnDropDesc')}</small>
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='TorrentsSavePath'
                    label={t('TorrentsSavePath')}
                    value={TorrentsSavePath}
                    type='url'
                    fullWidth
                  />
                </CacheStorageSettings>
              )}
            </MainSettingsContent>
          </TabPanel>

          <TabPanel value={selectedTab} index={1} dir={direction}>
            <SecondarySettingsContent>
              <SettingSectionLabel>Дополнительные настройки</SettingSectionLabel>

              <FormControlLabel
                control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='primary' />}
                label={t('EnableIPv6')}
              />
              <br />
              <FormControlLabel
                control={<Switch checked={!DisableTCP} onChange={inputForm} id='DisableTCP' color='primary' />}
                label={t('TCP')}
              />
              <br />
              <FormControlLabel
                control={<Switch checked={!DisableUTP} onChange={inputForm} id='DisableUTP' color='primary' />}
                label={t('UTP')}
              />
              <br />
              <FormControlLabel
                control={<Switch checked={!DisablePEX} onChange={inputForm} id='DisablePEX' color='primary' />}
                label={t('PEX')}
              />
              <br />
              <FormControlLabel
                control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='primary' />}
                label={t('ForceEncrypt')}
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='TorrentDisconnectTimeout'
                label={t('TorrentDisconnectTimeout')}
                value={TorrentDisconnectTimeout}
                type='number'
                fullWidth
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='ConnectionsLimit'
                label={t('ConnectionsLimit')}
                value={ConnectionsLimit}
                type='number'
                fullWidth
              />
              <br />
              <FormControlLabel
                control={<Switch checked={!DisableDHT} onChange={inputForm} id='DisableDHT' color='primary' />}
                label={t('DHT')}
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='DhtConnectionLimit'
                label={t('DhtConnectionLimit')}
                value={DhtConnectionLimit}
                type='number'
                fullWidth
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='DownloadRateLimit'
                label={t('DownloadRateLimit')}
                value={DownloadRateLimit}
                type='number'
                fullWidth
              />
              <br />
              <FormControlLabel
                control={<Switch checked={!DisableUpload} onChange={inputForm} id='DisableUpload' color='primary' />}
                label={t('Upload')}
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='UploadRateLimit'
                label={t('UploadRateLimit')}
                value={UploadRateLimit}
                type='number'
                fullWidth
              />
              <br />
              <TextField
                onChange={inputForm}
                margin='dense'
                id='PeersListenPort'
                label={t('PeersListenPort')}
                value={PeersListenPort}
                type='number'
                fullWidth
              />
              <br />
              <FormControlLabel
                control={<Switch checked={!DisableUPNP} onChange={inputForm} id='DisableUPNP' color='primary' />}
                label={t('UPNP')}
              />
              <br />
              <InputLabel htmlFor='RetrackersMode'>{t('RetrackersMode')}</InputLabel>
              <Select onChange={inputForm} type='number' native id='RetrackersMode' value={RetrackersMode}>
                <option value={0}>{t('DontAddRetrackers')}</option>
                <option value={1}>{t('AddRetrackers')}</option>
                <option value={2}>{t('RemoveRetrackers')}</option>
                <option value={3}>{t('ReplaceRetrackers')}</option>
              </Select>
              <br />
            </SecondarySettingsContent>
          </TabPanel>
        </SwipeableViews>
      </>
      {/* <DialogTitle id='form-dialog-title'>{t('Settings')}</DialogTitle>
      <DialogContent>
        <TextField
          onChange={onInputHost}
          margin='dense'
          id='TorrServerHost'
          label={t('Host')}
          value={tsHost}
          type='url'
          fullWidth
        />
        {show && (
          <>
            <TextField
              onChange={inputForm}
              margin='dense'
              id='CacheSize'
              label={t('CacheSize')}
              value={CacheSize}
              type='number'
              fullWidth
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='ReaderReadAHead'
              label={t('ReaderReadAHead')}
              value={ReaderReadAHead}
              type='number'
              fullWidth
            />
            <br />
            <FormControlLabel
              control={<Switch checked={PreloadBuffer} onChange={inputForm} id='PreloadBuffer' color='primary' />}
              label={t('PreloadBuffer')}
            />
            <br />
            <FormControlLabel
              control={<Switch checked={UseDisk} onChange={inputForm} id='UseDisk' color='primary' />}
              label={t('UseDisk')}
            />
            <br />
            <small>{t('UseDiskDesc')}</small>
            <br />
            <FormControlLabel
              control={
                <Switch checked={RemoveCacheOnDrop} onChange={inputForm} id='RemoveCacheOnDrop' color='primary' />
              }
              label={t('RemoveCacheOnDrop')}
            />
            <br />
            <small>{t('RemoveCacheOnDropDesc')}</small>
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='TorrentsSavePath'
              label={t('TorrentsSavePath')}
              value={TorrentsSavePath}
              type='url'
              fullWidth
            />
            <br />
            <FormControlLabel
              control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='primary' />}
              label={t('EnableIPv6')}
            />
            <br />
            <FormControlLabel
              control={<Switch checked={!DisableTCP} onChange={inputForm} id='DisableTCP' color='primary' />}
              label={t('TCP')}
            />
            <br />
            <FormControlLabel
              control={<Switch checked={!DisableUTP} onChange={inputForm} id='DisableUTP' color='primary' />}
              label={t('UTP')}
            />
            <br />
            <FormControlLabel
              control={<Switch checked={!DisablePEX} onChange={inputForm} id='DisablePEX' color='primary' />}
              label={t('PEX')}
            />
            <br />
            <FormControlLabel
              control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='primary' />}
              label={t('ForceEncrypt')}
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='TorrentDisconnectTimeout'
              label={t('TorrentDisconnectTimeout')}
              value={TorrentDisconnectTimeout}
              type='number'
              fullWidth
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='ConnectionsLimit'
              label={t('ConnectionsLimit')}
              value={ConnectionsLimit}
              type='number'
              fullWidth
            />
            <br />
            <FormControlLabel
              control={<Switch checked={!DisableDHT} onChange={inputForm} id='DisableDHT' color='primary' />}
              label={t('DHT')}
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='DhtConnectionLimit'
              label={t('DhtConnectionLimit')}
              value={DhtConnectionLimit}
              type='number'
              fullWidth
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='DownloadRateLimit'
              label={t('DownloadRateLimit')}
              value={DownloadRateLimit}
              type='number'
              fullWidth
            />
            <br />
            <FormControlLabel
              control={<Switch checked={!DisableUpload} onChange={inputForm} id='DisableUpload' color='primary' />}
              label={t('Upload')}
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='UploadRateLimit'
              label={t('UploadRateLimit')}
              value={UploadRateLimit}
              type='number'
              fullWidth
            />
            <br />
            <TextField
              onChange={inputForm}
              margin='dense'
              id='PeersListenPort'
              label={t('PeersListenPort')}
              value={PeersListenPort}
              type='number'
              fullWidth
            />
            <br />
            <FormControlLabel
              control={<Switch checked={!DisableUPNP} onChange={inputForm} id='DisableUPNP' color='primary' />}
              label={t('UPNP')}
            />
            <br />
            <InputLabel htmlFor='RetrackersMode'>{t('RetrackersMode')}</InputLabel>
            <Select onChange={inputForm} type='number' native id='RetrackersMode' value={RetrackersMode}>
              <option value={0}>{t('DontAddRetrackers')}</option>
              <option value={1}>{t('AddRetrackers')}</option>
              <option value={2}>{t('RemoveRetrackers')}</option>
              <option value={3}>{t('ReplaceRetrackers')}</option>
            </Select>
            <br />
          </>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} color='primary' variant='outlined'>
          {t('Cancel')}
        </Button>

        <Button onClick={handleSave} color='primary' variant='outlined'>
          {t('Save')}
        </Button>
      </DialogActions> */}
      <FooterSection>
        <FormControlLabel
          control={
            <Checkbox
              checked={isProMode}
              onChange={({ target: { checked } }) => {
                setIsProMode(checked)
                localStorage.setItem('isProMode', checked)
                if (!checked) setSelectedTab(0)
              }}
              color='primary'
            />
          }
          label='Pro mode'
        />

        <div>
          <Button onClick={handleClose} color='secondary' variant='outlined'>
            {t('Cancel')}
          </Button>

          <Button variant='contained' onClick={handleSave} color='primary'>
            {t('Save')}
          </Button>
        </div>
      </FooterSection>
    </Dialog>
  )
}
