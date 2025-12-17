import axios from 'axios'
import Button from '@material-ui/core/Button'
import Switch from '@material-ui/core/Switch'
import { FormControlLabel, useMediaQuery, useTheme } from '@material-ui/core'
import { settingsHost } from 'utils/Hosts'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import AppBar from '@material-ui/core/AppBar'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import SwipeableViews from 'react-swipeable-views'
import CircularProgress from '@material-ui/core/CircularProgress'
import { StyledDialog } from 'style/CustomMaterialUiStyles'
import useOnStandaloneAppOutsideClick from 'utils/useOnStandaloneAppOutsideClick'
import { isStandaloneApp } from 'utils/Utils'

import { SettingsHeader, FooterSection, Content } from './style'
import defaultSettings from './defaultSettings'
import { a11yProps, TabPanel } from './tabComponents'
import PrimarySettingsComponent from './PrimarySettingsComponent'
import SecondarySettingsComponent from './SecondarySettingsComponent'
import MobileAppSettings from './MobileAppSettings'
import TorznabSettings from './TorznabSettings'

export default function SettingsDialog({ handleClose }) {
  const { t } = useTranslation()
  const fullScreen = useMediaQuery('@media (max-width:930px)')
  const { direction } = useTheme()

  const [settings, setSettings] = useState()
  const [selectedTab, setSelectedTab] = useState(0)
  const [cacheSize, setCacheSize] = useState(32)
  const [cachePercentage, setCachePercentage] = useState(40)
  const [preloadCachePercentage, setPreloadCachePercentage] = useState(0)
  const [isProMode, setIsProMode] = useState(JSON.parse(localStorage.getItem('isProMode')) || false)
  const [isVlcUsed, setIsVlcUsed] = useState(JSON.parse(localStorage.getItem('isVlcUsed')) ?? false)
  const [isInfuseUsed, setIsInfuseUsed] = useState(JSON.parse(localStorage.getItem('isInfuseUsed')) ?? false)

  useEffect(() => {
    axios.post(settingsHost(), { action: 'get' }).then(({ data }) => {
      setSettings({ ...data, CacheSize: data.CacheSize / (1024 * 1024) })
    })
  }, [])

  const ref = useOnStandaloneAppOutsideClick(handleClose)

  const handleSave = () => {
    handleClose()
    const sets = JSON.parse(JSON.stringify(settings))
    sets.CacheSize = cacheSize * 1024 * 1024
    sets.ReaderReadAHead = cachePercentage
    sets.PreloadCache = preloadCachePercentage
    axios.post(settingsHost(), { action: 'set', sets })
    localStorage.setItem('isVlcUsed', isVlcUsed)
    localStorage.setItem('isInfuseUsed', isInfuseUsed)
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
    } else if (type === 'url' || type === 'text') {
      sets[id] = value
    }
    setSettings(sets)
  }

  const { CacheSize, ReaderReadAHead, PreloadCache } = settings || {}

  useEffect(() => {
    if (isNaN(CacheSize) || isNaN(ReaderReadAHead) || isNaN(PreloadCache)) return

    setCacheSize(CacheSize)
    setCachePercentage(ReaderReadAHead)
    setPreloadCachePercentage(PreloadCache)
  }, [CacheSize, ReaderReadAHead, PreloadCache])

  const updateSettings = newProps => setSettings({ ...settings, ...newProps })
  const handleChange = (_, newValue) => setSelectedTab(newValue)
  const handleChangeIndex = index => setSelectedTab(index)

  return (
    <StyledDialog open onClose={handleClose} fullScreen={fullScreen} fullWidth maxWidth='md' ref={ref}>
      <SettingsHeader>
        <div>{t('SettingsDialog.Settings')}</div>
        <FormControlLabel
          control={
            <Switch
              checked={isProMode}
              onChange={({ target: { checked } }) => {
                setIsProMode(checked)
                localStorage.setItem('isProMode', checked)
                if (!checked) setSelectedTab(0)
              }}
              style={{ color: 'white' }}
            />
          }
          label={t('SettingsDialog.ProMode')}
          labelPlacement='start'
        />
      </SettingsHeader>

      <AppBar position='static' color='default'>
        <Tabs
          value={selectedTab}
          onChange={handleChange}
          indicatorColor='secondary'
          textColor='secondary'
          variant='fullWidth'
        >
          <Tab label={t('SettingsDialog.Tabs.Main')} {...a11yProps(0)} />

          <Tab label='Torznab' {...a11yProps(1)} />

          <Tab
            disabled={!isProMode}
            label={
              <>
                <div>{t('SettingsDialog.Tabs.Additional')}</div>
                {!isProMode && <div style={{ fontSize: '9px' }}>{t('SettingsDialog.Tabs.AdditionalDisabled')}</div>}
              </>
            }
            {...a11yProps(2)}
          />

          {isStandaloneApp && <Tab label={t('SettingsDialog.Tabs.App')} {...a11yProps(3)} />}
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
                <PrimarySettingsComponent
                  settings={settings}
                  inputForm={inputForm}
                  cachePercentage={cachePercentage}
                  preloadCachePercentage={preloadCachePercentage}
                  cacheSize={cacheSize}
                  isProMode={isProMode}
                  setCacheSize={setCacheSize}
                  setCachePercentage={setCachePercentage}
                  setPreloadCachePercentage={setPreloadCachePercentage}
                  updateSettings={updateSettings}
                />
              </TabPanel>

              <TabPanel value={selectedTab} index={1} dir={direction}>
                <TorznabSettings settings={settings} inputForm={inputForm} updateSettings={updateSettings} />
              </TabPanel>

              <TabPanel value={selectedTab} index={2} dir={direction}>
                <SecondarySettingsComponent settings={settings} inputForm={inputForm} updateSettings={updateSettings} />
              </TabPanel>

              {isStandaloneApp && (
                <TabPanel value={selectedTab} index={3} dir={direction}>
                  <MobileAppSettings
                    isVlcUsed={isVlcUsed}
                    setIsVlcUsed={setIsVlcUsed}
                    isInfuseUsed={isInfuseUsed}
                    setIsInfuseUsed={setIsInfuseUsed}
                  />
                </TabPanel>
              )}
            </SwipeableViews>
          </>
        ) : (
          <CircularProgress color='secondary' />
        )}
      </Content>

      <FooterSection>
        <Button onClick={handleClose} color='secondary' variant='outlined'>
          {t('Cancel')}
        </Button>

        <Button
          onClick={() => {
            setCacheSize(defaultSettings.CacheSize)
            setCachePercentage(defaultSettings.ReaderReadAHead)
            setPreloadCachePercentage(defaultSettings.PreloadCache)
            updateSettings(defaultSettings)
          }}
          color='secondary'
          variant='outlined'
        >
          {t('SettingsDialog.ResetToDefault')}
        </Button>

        <Button variant='contained' onClick={handleSave} color='secondary'>
          {t('Save')}
        </Button>
      </FooterSection>
    </StyledDialog>
  )
}
