import axios from 'axios'
import Dialog from '@material-ui/core/Dialog'
import TextField from '@material-ui/core/TextField'
import Button from '@material-ui/core/Button'
import Checkbox from '@material-ui/core/Checkbox'
import {
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
import { settingsHost } from 'utils/Hosts'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Header } from 'style/DialogStyles'
import AppBar from '@material-ui/core/AppBar'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import SwipeableViews from 'react-swipeable-views'
import { USBIcon, RAMIcon } from 'icons'
import CircularProgress from '@material-ui/core/CircularProgress'

import {
  FooterSection,
  Divider,
  PreloadCacheValue,
  MainSettingsContent,
  SecondarySettingsContent,
  StorageButton,
  StorageIconWrapper,
  CacheStorageSelector,
  SettingSectionLabel,
  PreloadCachePercentage,
  cacheBeforeReaderColor,
  cacheAfterReaderColor,
  Content,
} from './style'
import defaultSettings from './defaultSettings'
import { a11yProps, TabPanel } from './tabComponents'

const CacheStorageLocationLabel = ({ style }) => {
  const { t } = useTranslation()

  return (
    <SettingSectionLabel style={style}>
      {t('SettingsDialog.CacheStorageLocation')}
      <small>{t('SettingsDialog.UseDiskDesc')}</small>
    </SettingSectionLabel>
  )
}

const SliderInput = ({
  isProMode,
  title,
  value,
  setValue,
  sliderMin,
  sliderMax,
  inputMin,
  inputMax,
  step = 1,
  onBlurCallback,
}) => {
  const onBlur = ({ target: { value } }) => {
    if (value < inputMin) return setValue(inputMin)
    if (value > inputMax) return setValue(inputMax)

    onBlurCallback && onBlurCallback(value)
  }

  const onInputChange = ({ target: { value } }) => setValue(value === '' ? '' : Number(value))
  const onSliderChange = (_, newValue) => setValue(newValue)

  return (
    <>
      <div>{title}</div>

      <Grid container spacing={2} alignItems='center'>
        <Grid item xs>
          <Slider min={sliderMin} max={sliderMax} value={value} onChange={onSliderChange} step={step} />
        </Grid>

        {isProMode && (
          <Grid item>
            <Input
              value={value}
              margin='dense'
              onChange={onInputChange}
              onBlur={onBlur}
              style={{ width: '65px' }}
              inputProps={{ step, min: inputMin, max: inputMax, type: 'number' }}
            />
          </Grid>
        )}
      </Grid>
    </>
  )
}

export default function SettingsDialog({ handleClose }) {
  const { t } = useTranslation()
  const fullScreen = useMediaQuery('@media (max-width:930px)')
  const { direction } = useTheme()

  const [settings, setSettings] = useState()
  const [selectedTab, setSelectedTab] = useState(0)
  const [cacheSize, setCacheSize] = useState(32)
  const [cachePercentage, setCachePercentage] = useState(40)
  const [isProMode, setIsProMode] = useState(JSON.parse(localStorage.getItem('isProMode')) || false)

  useEffect(() => {
    axios.post(settingsHost(), { action: 'get' }).then(({ data }) => {
      setSettings({ ...data, CacheSize: data.CacheSize / (1024 * 1024) })
    })
  }, [])

  const handleSave = () => {
    handleClose()
    const sets = JSON.parse(JSON.stringify(settings))
    sets.CacheSize = cacheSize * 1024 * 1024
    sets.ReaderReadAHead = cachePercentage
    axios.post(settingsHost(), { action: 'set', sets })
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
    setSettings(sets)
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
  } = settings || {}

  useEffect(() => {
    if (!CacheSize || !ReaderReadAHead) return

    setCacheSize(CacheSize)
    setCachePercentage(ReaderReadAHead)
  }, [CacheSize, ReaderReadAHead])

  const updateSettings = newProps => setSettings({ ...settings, ...newProps })
  const handleChange = (_, newValue) => setSelectedTab(newValue)
  const handleChangeIndex = index => setSelectedTab(index)

  return (
    <Dialog open onClose={handleClose} fullScreen={fullScreen} fullWidth maxWidth='md'>
      <Header>{t('SettingsDialog.Settings')}</Header>

      <AppBar position='static' color='default'>
        <Tabs
          value={selectedTab}
          onChange={handleChange}
          indicatorColor='primary'
          textColor='primary'
          variant='fullWidth'
        >
          <Tab label={t('SettingsDialog.Tabs.Main')} {...a11yProps(0)} />

          <Tab
            disabled={!isProMode}
            label={isProMode ? t('SettingsDialog.Tabs.Additional') : t('SettingsDialog.Tabs.AdditionalDisabled')}
            {...a11yProps(1)}
          />
        </Tabs>
      </AppBar>

      <Content isLoading={!settings}>
        {settings ? (
          <>
            <SwipeableViews
              axis={direction === 'rtl' ? 'x-reverse' : 'x'}
              index={selectedTab}
              onChangeIndex={handleChangeIndex}
            >
              <TabPanel value={selectedTab} index={0} dir={direction}>
                <MainSettingsContent>
                  <div>
                    <SettingSectionLabel>{t('SettingsDialog.CacheSettings')}</SettingSectionLabel>

                    <PreloadCachePercentage
                      value={100 - cachePercentage}
                      label={`${t('Cache')} ${cacheSize} MB`}
                      isPreloadEnabled={PreloadBuffer}
                    />

                    <PreloadCacheValue color={cacheBeforeReaderColor}>
                      <div>
                        {100 - cachePercentage}% ({Math.round((cacheSize / 100) * (100 - cachePercentage))} MB)
                      </div>

                      <div>{t('SettingsDialog.CacheBeforeReaderDesc')}</div>
                    </PreloadCacheValue>

                    <PreloadCacheValue color={cacheAfterReaderColor}>
                      <div>
                        {cachePercentage}% ({Math.round((cacheSize / 100) * cachePercentage)} MB)
                      </div>

                      <div>{t('SettingsDialog.CacheAfterReaderDesc')}</div>
                    </PreloadCacheValue>

                    <Divider />

                    <SliderInput
                      isProMode={isProMode}
                      title={t('SettingsDialog.CacheSize')}
                      value={cacheSize}
                      setValue={setCacheSize}
                      sliderMin={32}
                      sliderMax={1024}
                      inputMin={32}
                      inputMax={20000}
                      step={8}
                      onBlurCallback={value => setCacheSize(Math.round(value / 8) * 8)}
                    />

                    <SliderInput
                      isProMode={isProMode}
                      title={t('SettingsDialog.ReaderReadAHead')}
                      value={cachePercentage}
                      setValue={setCachePercentage}
                      sliderMin={40}
                      sliderMax={95}
                      inputMin={0}
                      inputMax={100}
                    />

                    <FormControlLabel
                      control={
                        <Switch checked={!!PreloadBuffer} onChange={inputForm} id='PreloadBuffer' color='primary' />
                      }
                      label={t('SettingsDialog.PreloadBuffer')}
                    />
                  </div>

                  {UseDisk ? (
                    <div>
                      <CacheStorageLocationLabel />

                      <div style={{ display: 'grid', gridAutoFlow: 'column' }}>
                        <StorageButton small onClick={() => updateSettings({ UseDisk: false })}>
                          <StorageIconWrapper small>
                            <RAMIcon color='#323637' />
                          </StorageIconWrapper>

                          <div>{t('SettingsDialog.RAM')}</div>
                        </StorageButton>

                        <StorageButton small selected>
                          <StorageIconWrapper small selected>
                            <USBIcon color='#dee3e5' />
                          </StorageIconWrapper>

                          <div>{t('SettingsDialog.Disk')}</div>
                        </StorageButton>
                      </div>

                      <FormControlLabel
                        control={
                          <Switch
                            checked={RemoveCacheOnDrop}
                            onChange={inputForm}
                            id='RemoveCacheOnDrop'
                            color='primary'
                          />
                        }
                        label={t('SettingsDialog.RemoveCacheOnDrop')}
                      />
                      <div>
                        <small>{t('SettingsDialog.RemoveCacheOnDropDesc')}</small>
                      </div>

                      <TextField
                        onChange={inputForm}
                        margin='dense'
                        id='TorrentsSavePath'
                        label={t('SettingsDialog.TorrentsSavePath')}
                        value={TorrentsSavePath}
                        type='url'
                        fullWidth
                      />
                    </div>
                  ) : (
                    <CacheStorageSelector>
                      <CacheStorageLocationLabel style={{ placeSelf: 'start', gridArea: 'label' }} />

                      <StorageButton selected>
                        <StorageIconWrapper selected>
                          <RAMIcon color='#dee3e5' />
                        </StorageIconWrapper>

                        <div>{t('SettingsDialog.RAM')}</div>
                      </StorageButton>

                      <StorageButton onClick={() => updateSettings({ UseDisk: true })}>
                        <StorageIconWrapper>
                          <USBIcon color='#323637' />
                        </StorageIconWrapper>

                        <div>{t('SettingsDialog.Disk')}</div>
                      </StorageButton>
                    </CacheStorageSelector>
                  )}
                </MainSettingsContent>
              </TabPanel>

              <TabPanel value={selectedTab} index={1} dir={direction}>
                <SecondarySettingsContent>
                  <SettingSectionLabel>{t('SettingsDialog.AdditionalSettings')}</SettingSectionLabel>

                  <FormControlLabel
                    control={<Switch checked={EnableIPv6} onChange={inputForm} id='EnableIPv6' color='primary' />}
                    label='IPv6'
                  />
                  <br />
                  <FormControlLabel
                    control={<Switch checked={!DisableTCP} onChange={inputForm} id='DisableTCP' color='primary' />}
                    label='TCP (Transmission Control Protocol)'
                  />
                  <br />
                  <FormControlLabel
                    control={<Switch checked={!DisableUTP} onChange={inputForm} id='DisableUTP' color='primary' />}
                    label='μTP (Micro Transport Protocol)'
                  />
                  <br />
                  <FormControlLabel
                    control={<Switch checked={!DisablePEX} onChange={inputForm} id='DisablePEX' color='primary' />}
                    label='PEX (Peer Exchange)'
                  />
                  <br />
                  <FormControlLabel
                    control={<Switch checked={ForceEncrypt} onChange={inputForm} id='ForceEncrypt' color='primary' />}
                    label={t('SettingsDialog.ForceEncrypt')}
                  />
                  <br />
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='TorrentDisconnectTimeout'
                    label={t('SettingsDialog.TorrentDisconnectTimeout')}
                    value={TorrentDisconnectTimeout}
                    type='number'
                    fullWidth
                  />
                  <br />
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='ConnectionsLimit'
                    label={t('SettingsDialog.ConnectionsLimit')}
                    value={ConnectionsLimit}
                    type='number'
                    fullWidth
                  />
                  <br />
                  <FormControlLabel
                    control={<Switch checked={!DisableDHT} onChange={inputForm} id='DisableDHT' color='primary' />}
                    label={t('SettingsDialog.DHT')}
                  />
                  <br />
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='DhtConnectionLimit'
                    label={t('SettingsDialog.DhtConnectionLimit')}
                    value={DhtConnectionLimit}
                    type='number'
                    fullWidth
                  />
                  <br />
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='DownloadRateLimit'
                    label={t('SettingsDialog.DownloadRateLimit')}
                    value={DownloadRateLimit}
                    type='number'
                    fullWidth
                  />
                  <br />
                  <FormControlLabel
                    control={
                      <Switch checked={!DisableUpload} onChange={inputForm} id='DisableUpload' color='primary' />
                    }
                    label={t('SettingsDialog.Upload')}
                  />
                  <br />
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='UploadRateLimit'
                    label={t('SettingsDialog.UploadRateLimit')}
                    value={UploadRateLimit}
                    type='number'
                    fullWidth
                  />
                  <br />
                  <TextField
                    onChange={inputForm}
                    margin='dense'
                    id='PeersListenPort'
                    label={t('SettingsDialog.PeersListenPort')}
                    value={PeersListenPort}
                    type='number'
                    fullWidth
                  />
                  <br />
                  <FormControlLabel
                    control={<Switch checked={!DisableUPNP} onChange={inputForm} id='DisableUPNP' color='primary' />}
                    label='UPnP (Universal Plug and Play)'
                  />
                  <br />
                  <InputLabel htmlFor='RetrackersMode'>{t('SettingsDialog.RetrackersMode')}</InputLabel>
                  <Select onChange={inputForm} type='number' native id='RetrackersMode' value={RetrackersMode}>
                    <option value={0}>{t('SettingsDialog.DontAddRetrackers')}</option>
                    <option value={1}>{t('SettingsDialog.AddRetrackers')}</option>
                    <option value={2}>{t('SettingsDialog.RemoveRetrackers')}</option>
                    <option value={3}>{t('SettingsDialog.ReplaceRetrackers')}</option>
                  </Select>
                  <br />
                </SecondarySettingsContent>
              </TabPanel>
            </SwipeableViews>
          </>
        ) : (
          <CircularProgress color='secondary' />
        )}
      </Content>

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
          label={t('SettingsDialog.ProMode')}
        />

        <div>
          <Button onClick={handleClose} color='secondary' variant='outlined'>
            {t('Cancel')}
          </Button>

          <Button
            onClick={() => {
              setCacheSize(defaultSettings.CacheSize)
              setCachePercentage(defaultSettings.ReaderReadAHead)
              updateSettings(defaultSettings)
            }}
            color='secondary'
            variant='outlined'
          >
            Reset to default
          </Button>

          <Button variant='contained' onClick={handleSave} color='primary'>
            {t('Save')}
          </Button>
        </div>
      </FooterSection>
    </Dialog>
  )
}
