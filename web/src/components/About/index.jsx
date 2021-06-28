import { useState } from 'react'
import Button from '@material-ui/core/Button'
import Dialog from '@material-ui/core/Dialog'
import InfoIcon from '@material-ui/icons/Info'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useTranslation } from 'react-i18next'
import { GitHub as GitHubIcon } from '@material-ui/icons'
import { useMediaQuery } from '@material-ui/core'

import NameComponent from './NameComponent'
import tsIcon from './ts-icon-192x192.png'
import { DialogWrapper, HeaderSection, ThanksSection, SpecialThanksSection, FooterSection } from './style'

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

            <SpecialThanksSection>
              <span>{t('SpecialThanks')}</span>

              <div>
                <NameComponent name='Daniel Shleifman' link='https://github.com/dancheskus' />
                <NameComponent name='Matt Joiner' link='https://github.com/anacrolix' />
                <NameComponent name='nikk' link='https://github.com/tsynik' />
                <NameComponent name='tw1cker Руслан Пахнев' link='https://github.com/Nemiroff' />
                <NameComponent name='SpAwN_LMG' link='https://github.com/spawnlmg' />
              </div>
            </SpecialThanksSection>
          </div>

          <FooterSection>
            <a href='https://github.com/YouROK/TorrServer' target='_blank' rel='noreferrer'>
              <Button color='primary' variant='outlined'>
                <GitHubIcon style={{ marginRight: '10px' }} />
                {t('ProjectSource')}
              </Button>
            </a>

            <Button onClick={() => setOpen(false)} color='primary' variant='contained' autoFocus>
              {t('Close')}
            </Button>
          </FooterSection>
        </DialogWrapper>
      </Dialog>
    </>
  )
}
