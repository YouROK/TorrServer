import { useState } from 'react'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import InfoIcon from '@material-ui/icons/Info'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useTranslation } from 'react-i18next'
import { useMediaQuery } from '@material-ui/core'

import LinkComponent from './LinkComponent'
import tsIcon from './ts-icon-192x192.png'
import { DialogWrapper, HeaderSection, ThanksSection, Section, FooterSection } from './style'

export default function AboutDialog() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const fullScreen = useMediaQuery('@media (max-width:930px)')

  return (
    <>
      <ListItem button key='Settings' onClick={() => setOpen(true)}>
        <ListItemIcon>
          <InfoIcon />
        </ListItemIcon>
        <ListItemText primary={t('About')} />
      </ListItem>

      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        aria-labelledby='form-dialog-title'
        fullScreen={fullScreen}
        maxWidth='xl'
      >
        <DialogWrapper>
          <HeaderSection>
            <div>{t('About')}</div>
            <img src={tsIcon} alt='ts-icon' />
          </HeaderSection>

          <div style={{ overflow: 'auto' }}>
            <ThanksSection>{t('ThanksToEveryone')}</ThanksSection>

            <Section>
              <span>{t('Links')}</span>

              <div>
                <LinkComponent name={t('ProjectSource')} link='https://github.com/YouROK/TorrServer' />
                <LinkComponent name={t('Releases')} link='https://github.com/YouROK/TorrServer/releases' />
                <LinkComponent name='API' link='https://github.com/YouROK/TorrServer/blob/master/README.md' />
              </div>
            </Section>

            <Section>
              <span>{t('SpecialThanks')}</span>

              <div>
                <LinkComponent name='Daniel Shleifman' link='https://github.com/dancheskus' />
                <LinkComponent name='Matt Joiner' link='https://github.com/anacrolix' />
                <LinkComponent name='nikk' link='https://github.com/tsynik' />
                <LinkComponent name='tw1cker Руслан Пахнев' link='https://github.com/Nemiroff' />
                <LinkComponent name='SpAwN_LMG' link='https://github.com/spawnlmg' />
              </div>
            </Section>
          </div>

          <FooterSection>
            <Button onClick={() => setOpen(false)} color='primary' variant='contained' autoFocus>
              {t('Close')}
            </Button>
          </FooterSection>
        </DialogWrapper>
      </Dialog>
    </>
  )
}
