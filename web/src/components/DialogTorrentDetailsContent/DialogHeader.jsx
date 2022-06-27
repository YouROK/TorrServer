import { AppBar, IconButton, makeStyles, Toolbar, Typography } from '@material-ui/core'
import CloseIcon from '@material-ui/icons/Close'
import { ArrowBack } from '@material-ui/icons'
import { isStandaloneApp } from 'utils/Utils'

const useStyles = makeStyles({
  appBar: { position: 'relative', ...(isStandaloneApp && { paddingTop: '30px' }) },
  title: { marginLeft: '5px', flex: 1 },
})

export default function DialogHeader({ title, onClose, onBack }) {
  const classes = useStyles()

  return (
    <AppBar className={classes.appBar}>
      <Toolbar>
        {onBack && (
          <IconButton edge='start' color='inherit' onClick={onBack} aria-label='back'>
            <ArrowBack />
          </IconButton>
        )}

        <Typography variant='h6' className={classes.title}>
          {title}
        </Typography>

        <IconButton autoFocus color='inherit' onClick={onClose} aria-label='close' style={{ marginRight: '-10px' }}>
          <CloseIcon />
        </IconButton>
      </Toolbar>
    </AppBar>
  )
}
