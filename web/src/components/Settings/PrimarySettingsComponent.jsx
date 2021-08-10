import { useTranslation } from 'react-i18next'
import { USBIcon, RAMIcon } from 'icons'
import { FormControlLabel, Switch } from '@material-ui/core'
import TextField from '@material-ui/core/TextField'

import {
  PreloadCacheValue,
  MainSettingsContent,
  StorageButton,
  StorageIconWrapper,
  CacheStorageSelector,
  SettingSectionLabel,
  PreloadCachePercentage,
  cacheBeforeReaderColor,
  cacheAfterReaderColor,
} from './style'
import SliderInput from './SliderInput'

const CacheStorageLocationLabel = ({ style }) => {
  const { t } = useTranslation()

  return (
    <SettingSectionLabel style={style}>
      {t('SettingsDialog.CacheStorageLocation')}
      <small>{t('SettingsDialog.UseDiskDesc')}</small>
    </SettingSectionLabel>
  )
}

export default function PrimarySettingsComponent({
  settings,
  inputForm,
  cachePercentage,
  preloadCachePercentage,
  cacheSize,
  isProMode,
  setCacheSize,
  setCachePercentage,
  setPreloadCachePercentage,
  updateSettings,
}) {
  const { t } = useTranslation()
  const { UseDisk, TorrentsSavePath, RemoveCacheOnDrop } = settings || {}
  const preloadCacheSize = Math.round((cacheSize / 100) * preloadCachePercentage)

  return (
    <MainSettingsContent>
      <div>
        <SettingSectionLabel>{t('SettingsDialog.CacheSettings')}</SettingSectionLabel>

        <PreloadCachePercentage
          value={100 - cachePercentage}
          label={`${t('Cache')} ${cacheSize} ${t('MB')}`}
          preloadCachePercentage={preloadCachePercentage}
        />

        <PreloadCacheValue color={cacheBeforeReaderColor}>
          <div>
            {100 - cachePercentage}% ({Math.round((cacheSize / 100) * (100 - cachePercentage))} {t('MB')})
          </div>

          <div>{t('SettingsDialog.CacheBeforeReaderDesc')}</div>
        </PreloadCacheValue>

        <PreloadCacheValue color={cacheAfterReaderColor}>
          <div>
            {cachePercentage}% ({Math.round((cacheSize / 100) * cachePercentage)} {t('MB')})
          </div>

          <div>{t('SettingsDialog.CacheAfterReaderDesc')}</div>
        </PreloadCacheValue>

        <br />

        <SliderInput
          isProMode={isProMode}
          title={t('SettingsDialog.CacheSize')}
          value={cacheSize}
          setValue={setCacheSize}
          sliderMin={32}
          sliderMax={1024}
          inputMin={32}
          inputMax={999999}
          step={4}
          onBlurCallback={value => setCacheSize(Math.round(value / 4) * 4)}
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

        <SliderInput
          isProMode={isProMode}
          title={`${t('SettingsDialog.PreloadCache')} - ${preloadCachePercentage}% (${preloadCacheSize} ${t('MB')})`}
          value={preloadCachePercentage}
          setValue={setPreloadCachePercentage}
          sliderMin={0}
          sliderMax={100}
          inputMin={0}
          inputMax={100}
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
              <Switch checked={RemoveCacheOnDrop} onChange={inputForm} id='RemoveCacheOnDrop' color='secondary' />
            }
            label={t('SettingsDialog.RemoveCacheOnDrop')}
            labelPlacement='start'
          />
          <div>
            <small>{t('SettingsDialog.RemoveCacheOnDropDesc')}</small>
          </div>
          <br />
          <TextField
            onChange={inputForm}
            margin='normal'
            id='TorrentsSavePath'
            label={t('SettingsDialog.TorrentsSavePath')}
            value={TorrentsSavePath}
            type='url'
            variant='outlined'
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
  )
}
