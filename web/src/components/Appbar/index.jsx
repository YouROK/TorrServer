import { useEffect, useState } from 'react'
import clsx from 'clsx'
import { useTheme } from '@material-ui/core/styles'
import Drawer from '@material-ui/core/Drawer'
import AppBar from '@material-ui/core/AppBar'
import Toolbar from '@material-ui/core/Toolbar'
import List from '@material-ui/core/List'
import Typography from '@material-ui/core/Typography'
import Divider from '@material-ui/core/Divider'
import IconButton from '@material-ui/core/IconButton'
import MenuIcon from '@material-ui/icons/Menu'
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft'
import ChevronRightIcon from '@material-ui/icons/ChevronRight'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import CreditCardIcon from '@material-ui/icons/CreditCard'
import ListIcon from '@material-ui/icons/List'
import PowerSettingsNewIcon from '@material-ui/icons/PowerSettingsNew'
import { playlistAllHost, shutdownHost, getTorrServerHost } from 'utils/Hosts'
import TorrentList from 'components/TorrentList'
import AddDialogButton from 'components/Add'
import RemoveAll from 'components/RemoveAll'
import SettingsDialog from 'components/Settings'
import AboutDialog from 'components/About'
import DonateSnackbar from 'components/Donate'
import DonateDialog from 'components/Donate/DonateDialog'
import UploadDialog from 'components/Upload'

import useStyles from './useStyles'

export default function MiniDrawer() {
  const classes = useStyles()
  const theme = useTheme()
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [isDonationDialogOpen, setIsDonationDialogOpen] = useState(false)
  const [tsVersion, setTSVersion] = useState('')

  const handleDrawerOpen = () => {
    setIsDrawerOpen(true)
  }

  const handleDrawerClose = () => {
    setIsDrawerOpen(false)
  }

  useEffect(() => {
    fetch(`${getTorrServerHost()}/echo`)
      .then(resp => resp.text())
      .then(txt => {
        if (!txt.startsWith('<!DOCTYPE html>')) setTSVersion(txt)
      })
  }, [isDrawerOpen])

  return (
    <div className={classes.root}>
      <AppBar
        position='fixed'
        className={clsx(classes.appBar, {
          [classes.appBarShift]: isDrawerOpen,
        })}
      >
        <Toolbar>
          <IconButton
            color='inherit'
            aria-label='open drawer'
            onClick={handleDrawerOpen}
            edge='start'
            className={clsx(classes.menuButton, {
              [classes.hide]: isDrawerOpen,
            })}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant='h6' noWrap>
            TorrServer {tsVersion}
          </Typography>
        </Toolbar>
      </AppBar>

      <Drawer
        variant='permanent'
        className={clsx(classes.drawer, {
          [classes.drawerOpen]: isDrawerOpen,
          [classes.drawerClose]: !isDrawerOpen,
        })}
        classes={{
          paper: clsx({
            [classes.drawerOpen]: isDrawerOpen,
            [classes.drawerClose]: !isDrawerOpen,
          }),
        }}
      >
        <div className={classes.toolbar}>
          <IconButton onClick={handleDrawerClose}>
            {theme.direction === 'rtl' ? <ChevronRightIcon /> : <ChevronLeftIcon />}
          </IconButton>
        </div>

        <Divider />

        <List>
          <AddDialogButton />
          <UploadDialog />
          <RemoveAll />
          <ListItem button component='a' key='Playlist all torrents' target='_blank' href={playlistAllHost()}>
            <ListItemIcon>
              <ListIcon />
            </ListItemIcon>
            <ListItemText primary='Playlist all torrents' />
          </ListItem>
        </List>

        <Divider />

        <List>
          <SettingsDialog />
          <AboutDialog />
          <ListItem button key='Close server' onClick={() => fetch(shutdownHost())}>
            <ListItemIcon>
              <PowerSettingsNewIcon />
            </ListItemIcon>
            <ListItemText primary='Close server' />
          </ListItem>
        </List>

        <Divider />

        <List>
          <ListItem button key='Donation' onClick={() => setIsDonationDialogOpen(true)}>
            <ListItemIcon>
              <CreditCardIcon />
            </ListItemIcon>
            <ListItemText primary='Donate' />
          </ListItem>
        </List>
      </Drawer>

      <main className={classes.content}>
        <div className={classes.toolbar} />

        <TorrentList />
      </main>

      {isDonationDialogOpen && <DonateDialog onClose={() => setIsDonationDialogOpen(false)} />}
      {!JSON.parse(localStorage.getItem('snackbarIsClosed')) && <DonateSnackbar />}
    </div>
  )
}
