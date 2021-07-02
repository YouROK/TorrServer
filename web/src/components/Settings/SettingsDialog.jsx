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
import { settingsHost } from 'utils/Hosts'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Header } from 'style/DialogStyles'
import AppBar from '@material-ui/core/AppBar'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import SwipeableViews from 'react-swipeable-views'
import { USBIcon, RAMIcon } from 'icons'

import {
  FooterSection,
  Divider,
  PreloadCacheValue,
  MainSettingsContent,
  SecondarySettingsContent,
  StorageButton,
  StorageIconWrapper,
  CacheSizeSettings,
  CacheStorageSelector,
  CacheStorageSettings,
  SettingSection,
  SettingLabel,
  SettingSectionLabel,
  PreloadCachePercentage,
  cacheBeforeReaderColor,
  cacheAfterReaderColor,
} from './style'

const a11yProps = index => ({
  id: `full-width-tab-${index}`,
  'aria-controls': `full-width-tabpanel-${index}`,
})

const TabPanel = ({ children, value, index, ...other }) => (
  <div role='tabpanel' hidden={value !== index} id={`full-width-tabpanel-${index}`} {...other}>
    {value === index && <>{children}</>}
  </div>
)

const defaultSettings = {
  CacheSize: 96,
  ReaderReadAHead: 95,
  UseDisk: false,
  UploadRateLimit: 0,
  TorrentsSavePath: '',
  ConnectionsLimit: 23,
  DhtConnectionLimit: 500,
  DisableDHT: false,
  DisablePEX: false,
  DisableTCP: false,
  DisableUPNP: false,
  DisableUTP: true,
  DisableUpload: false,
  DownloadRateLimit: 0,
  EnableDebug: false,
  EnableIPv6: false,
  ForceEncrypt: false,
  PeersListenPort: 0,
  PreloadBuffer: false,
  RemoveCacheOnDrop: false,
  RetrackersMode: 1,
  Strategy: 0,
  TorrentDisconnectTimeout: 30,
}

export default function SettingsDialog({ handleClose }) {
  const { t } = useTranslation()

  const fullScreen = useMediaQuery('@media (max-width:930px)')

  const [settings, setSettings] = useState()
  const [show, setShow] = useState(false)

  useEffect(() => {
    axios
      .post(settingsHost(), { action: 'get' })
      .then(({ data }) => {
        setSettings({ ...data, CacheSize: data.CacheSize / (1024 * 1024) })
        setShow(true)
      })
      .catch(() => setShow(false))
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

  const updateSettings = newProps => setSettings({ ...settings, ...newProps })

  const { direction } = useTheme()
  const [selectedTab, setSelectedTab] = useState(0)

  const handleChange = (_, newValue) => setSelectedTab(newValue)

  const handleChangeIndex = index => setSelectedTab(index)

  const [cacheSize, setCacheSize] = useState(32)
  const [cachePercentage, setCachePercentage] = useState(40)
  const [isProMode, setIsProMode] = useState(JSON.parse(localStorage.getItem('isProMode')) || false)

  useEffect(() => {
    if (!CacheSize || !ReaderReadAHead) return

    setCacheSize(CacheSize)
    setCachePercentage(ReaderReadAHead)
  }, [CacheSize, ReaderReadAHead])

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

      {settings ? (
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

                  <PreloadCacheValue color={cacheBeforeReaderColor}>
                    <div>
                      {100 - cachePercentage}% ({Math.round((cacheSize / 100) * (100 - cachePercentage))} МБ)
                    </div>
                    <div>От кеша будет оставаться позади воспроизводимого блока</div>
                  </PreloadCacheValue>

                  <PreloadCacheValue color={cacheAfterReaderColor}>
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
                        <Slider
                          min={32}
                          max={1024}
                          value={cacheSize}
                          onChange={(_, newValue) => setCacheSize(newValue)}
                          step={8}
                        />
                      </Grid>

                      {isProMode && (
                        <Grid item>
                          <Input
                            value={cacheSize}
                            margin='dense'
                            onChange={handleInputChange}
                            onBlur={handleBlur}
                            style={{ width: '65px' }}
                            inputProps={{ step: 8, min: 32, max: 20000, type: 'number' }}
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
                            inputProps={{ min: 0, max: 100, type: 'number' }}
                          />
                        </Grid>
                      )}
                    </Grid>
                  </SettingSection>

                  <FormControlLabel
                    control={
                      <Switch checked={!!PreloadBuffer} onChange={inputForm} id='PreloadBuffer' color='primary' />
                    }
                    label={t('PreloadBuffer')}
                  />
                </CacheSizeSettings>

                {UseDisk ? (
                  <CacheStorageSettings>
                    <SettingSectionLabel>Место хранения кеша</SettingSectionLabel>

                    <div style={{ display: 'grid', gridAutoFlow: 'column' }}>
                      <StorageButton small onClick={() => updateSettings({ UseDisk: false })}>
                        <StorageIconWrapper small>
                          <RAMIcon color='#323637' />
                        </StorageIconWrapper>

                        <div>Оперативная память</div>
                      </StorageButton>

                      <StorageButton small selected>
                        <StorageIconWrapper small selected>
                          <USBIcon color='#dee3e5' />
                        </StorageIconWrapper>

                        <div>Диск</div>
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
                ) : (
                  <CacheStorageSelector>
                    <SettingSectionLabel style={{ placeSelf: 'start', gridArea: 'label' }}>
                      Место хранения кеша
                    </SettingSectionLabel>

                    <StorageButton selected>
                      <StorageIconWrapper selected>
                        <RAMIcon color='#dee3e5' />
                      </StorageIconWrapper>

                      <div>Оперативная память</div>
                    </StorageButton>

                    <StorageButton onClick={() => updateSettings({ UseDisk: true })}>
                      <StorageIconWrapper>
                        <USBIcon color='#323637' />
                      </StorageIconWrapper>

                      <div>Диск</div>
                    </StorageButton>
                  </CacheStorageSelector>
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
      ) : (
        'loading...'
      )}
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
