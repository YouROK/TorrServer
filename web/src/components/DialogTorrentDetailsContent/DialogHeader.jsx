import Button from '@material-ui/core/Button'
import { AppBar, IconButton, makeStyles, Toolbar, Typography } from '@material-ui/core'
import CloseIcon from '@material-ui/icons/Close'
import { ArrowBack } from '@material-ui/icons'
import { useTranslation } from 'react-i18next'

const useStyles = makeStyles(theme => ({
  appBar: { position: 'relative' },
  title: { marginLeft: theme.spacing(2), flex: 1 },
}))

export default function DialogHeader({ title, onClose, onBack }) {
  const { t } = useTranslation()
  const classes = useStyles()

  return (
    <AppBar className={classes.appBar}>
      <Toolbar>
        <IconButton edge='start' color='inherit' onClick={onBack || onClose} aria-label='close'>
          {onBack ? <ArrowBack /> : <CloseIcon />}
        </IconButton>

        <Typography variant='h6' className={classes.title}>
          {title}
        </Typography>

        {onBack && (
          <Button autoFocus color='inherit' onClick={onClose}>
            {t('Close')}
          </Button>
        )}
      </Toolbar>
    </AppBar>
  )
}
