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

import ListIcon from '@material-ui/icons/List'
import PowerSettingsNewIcon from '@material-ui/icons/PowerSettingsNew'

import TorrentList from '../TorrentList'

import AddDialogButton from '../Add'
import RemoveAll from '../RemoveAll'
import SettingsDialog from '../Settings'
import AboutDialog from '../About'
import { playlistAllHost, shutdownHost, torrserverHost } from '../../utils/Hosts'
import DonateDialog from '../Donate'
import UploadDialog from '../Upload'
import useStyles from './useStyles'

export default function MiniDrawer() {
    const classes = useStyles()
    const theme = useTheme()
    const [open, setOpen] = useState(false)
    const [tsVersion, setTSVersion] = useState('')

    const handleDrawerOpen = () => {
        setOpen(true)
    }

    const handleDrawerClose = () => {
        setOpen(false)
    }

    useEffect(() => {
        fetch(torrserverHost + '/echo')
            .then((resp) => resp.text())
            .then((txt) => {
                if (!txt.startsWith('<!DOCTYPE html>')) setTSVersion(txt)
            })
    }, [open])

    return (
        <div className={classes.root}>
            <AppBar
                position="fixed"
                className={clsx(classes.appBar, {
                    [classes.appBarShift]: open,
                })}
            >
                <Toolbar>
                    <IconButton
                        color="inherit"
                        aria-label="open drawer"
                        onClick={handleDrawerOpen}
                        edge="start"
                        className={clsx(classes.menuButton, {
                            [classes.hide]: open,
                        })}
                    >
                        <MenuIcon />
                    </IconButton>
                    <Typography variant="h6" noWrap>
                        TorrServer {tsVersion}
                    </Typography>
                </Toolbar>
            </AppBar>

            <Drawer
                variant="permanent"
                className={clsx(classes.drawer, {
                    [classes.drawerOpen]: open,
                    [classes.drawerClose]: !open,
                })}
                classes={{
                    paper: clsx({
                        [classes.drawerOpen]: open,
                        [classes.drawerClose]: !open,
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
                    <ListItem button component="a" key="Playlist all torrents" target="_blank" href={playlistAllHost()}>
                        <ListItemIcon>
                            <ListIcon />
                        </ListItemIcon>
                        <ListItemText primary="Playlist all torrents" />
                    </ListItem>
                </List>

                <Divider />

                <List>
                    <SettingsDialog />
                    <AboutDialog />
                    <ListItem button key="Close server" onClick={() => fetch(shutdownHost())}>
                        <ListItemIcon>
                            <PowerSettingsNewIcon />
                        </ListItemIcon>
                        <ListItemText primary="Close server" />
                    </ListItem>
                </List>
            </Drawer>
            
            <main className={classes.content}>
                <div className={classes.toolbar} />
                
                <TorrentList />
            </main>

            <DonateDialog />
        </div>
    )
}
