import axios from 'axios'
import { useEffect, useState } from 'react'
import Button from '@material-ui/core/Button'
import InfoIcon from '@material-ui/icons/Info'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useTranslation } from 'react-i18next'
import { useMediaQuery } from '@material-ui/core'
import { echoHost } from 'utils/Hosts'
import { StyledDialog, StyledMenuButtonWrapper } from 'style/CustomMaterialUiStyles'
import { isStandaloneApp } from 'utils/Utils'
import useOnStandaloneAppOutsideClick from 'utils/useOnStandaloneAppOutsideClick'

import LinkComponent from './LinkComponent'
import { DialogWrapper, HeaderSection, ThanksSection, Section, FooterSection } from './style'

export default function AboutDialog() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [torrServerVersion, setTorrServerVersion] = useState('')
  const fullScreen = useMediaQuery('@media (max-width:930px)')
  useEffect(() => {
    axios.get(echoHost()).then(({ data }) => setTorrServerVersion(data))
  }, [])

  const onClose = () => setOpen(false)
  const ref = useOnStandaloneAppOutsideClick(onClose)
  const getBasePath = () => {
    if (typeof window !== 'undefined') {
      return window.location.pathname.split('/')[1] || ''
    }
    return ''
  }
  const basePath = getBasePath()

  return (
    <>
      <StyledMenuButtonWrapper button key='Settings' onClick={() => setOpen(true)}>
        {isStandaloneApp ? (
          <>
            <InfoIcon />
            <div>{t('Details')}</div>
          </>
        ) : (
          <>
            <ListItemIcon>
              <InfoIcon />
            </ListItemIcon>

            <ListItemText primary={t('About')} />
          </>
        )}
      </StyledMenuButtonWrapper>

      <StyledDialog
        open={open}
        onClose={onClose}
        aria-labelledby='form-dialog-title'
        fullScreen={fullScreen}
        maxWidth='xl'
        ref={ref}
      >
        <DialogWrapper>
          <HeaderSection>
            <div>{t('About')}</div>
            {torrServerVersion}
            <img src={`${basePath}/icon.png`} alt='ts-icon' />
          </HeaderSection>

          <div style={{ overflow: 'auto' }}>
            <ThanksSection>{t('ThanksToEveryone')}</ThanksSection>

            <Section>
              <span>{t('Links')}</span>

              <div>
                <LinkComponent name={t('ProjectSource')} link='https://github.com/YouROK/TorrServer' />
                <LinkComponent name={t('Releases')} link='https://github.com/YouROK/TorrServer/releases' />
                <LinkComponent name={t('NasReleases')} link='https://github.com/vladlenas' />
                <LinkComponent name={t('ApiDocs')} link='swagger/index.html' />
              </div>
            </Section>

            <Section>
              <span>{t('SpecialThanks')}</span>

              <div>
                <LinkComponent name='Matt Joiner' link='https://github.com/anacrolix' />
                <LinkComponent name='Daniel Shleifman' link='https://github.com/dancheskus' />
                <LinkComponent name='nikk' link='https://github.com/tsynik' />
                <LinkComponent name='kolsys' link='https://github.com/kolsys' />
                <LinkComponent name='tw1cker' link='https://github.com/Nemiroff' />
                <LinkComponent name='SpAwN_LMG' link='https://github.com/spawnlmg' />
                <LinkComponent name='damiva' link='https://github.com/damiva' />
                <LinkComponent name='Vladlenas' link='https://github.com/vladlenas' />
                <LinkComponent name='Pavel Pikta' link='https://github.com/pavelpikta' />
                <LinkComponent name='Anton Potekhin' link='https://github.com/Anton111111' />
                <LinkComponent name='FaintGhost' link='https://github.com/FaintGhost' />
                <LinkComponent name='TopperBG' link='https://github.com/TopperBG' />
                <LinkComponent name='Evgeni' link='https://github.com/lieranderl' />
                <LinkComponent name='cocool97' link='https://github.com/cocool97' />
                <LinkComponent name='shadeov' link='https://github.com/shadeov' />
                <LinkComponent name='Pavel' link='https://github.com/butaford' />
                <LinkComponent name='Alexey Filimonov' link='https://github.com/filimonic' />
                <LinkComponent name='Viacheslav Evseev' link='https://github.com/leporel' />
              </div>
            </Section>
          </div>

          <FooterSection>
            <Button onClick={onClose} color='primary' variant='contained'>
              {t('Close')}
            </Button>
          </FooterSection>
        </DialogWrapper>
      </StyledDialog>
    </>
  )
}
