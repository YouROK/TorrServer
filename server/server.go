package server

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"server/log"
	"server/settings"
	"server/web"
)

func Start(port string, roSets bool) {
	settings.InitSets(roSets)
	if port == "" {
		port = "8090"
	}
	log.TLogln("Check web port", port)
	l, err := net.Listen("tcp", ":"+port)
	if l != nil {
		l.Close()
	}
	if err != nil {
		log.TLogln("Port", port, "already in use! Abort")
		os.Exit(1)
	} else {
	  go cleanCache()
	  web.Start(port)
	}
}

func cleanCache() {
	if !settings.BTsets.UseDisk || settings.BTsets.TorrentsSavePath == "/" || settings.BTsets.TorrentsSavePath == "" {
		return
	}

	dirs, err := ioutil.ReadDir(settings.BTsets.TorrentsSavePath)
	if err != nil {
		return
	}

	torrs := settings.ListTorrent()

	log.TLogln("Remove unused cache in dir:", settings.BTsets.TorrentsSavePath)
	for _, d := range dirs {
		if len(d.Name()) != 40 {
			// Not a hash
			continue
		}

		if !settings.BTsets.RemoveCacheOnDrop {
			for _, t := range torrs {
				if d.IsDir() && d.Name() != t.InfoHash.HexString() {
					log.TLogln("Remove unused cache:", d.Name())
					removeAllFiles(filepath.Join(settings.BTsets.TorrentsSavePath, d.Name()))
					break
				}
			}
		} else {
			if d.IsDir() {
				log.TLogln("Remove unused cache:", d.Name())
				removeAllFiles(filepath.Join(settings.BTsets.TorrentsSavePath, d.Name()))
			}
		}
	}
}

func removeAllFiles(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	for _, f := range files {
		name := filepath.Join(path, f.Name())
		os.Remove(name)
	}
	os.Remove(path)
}

func WaitServer() string {
	err := web.Wait()
	if err != nil {
		return err.Error()
	}
	return ""
}

func Stop() {
	web.Stop()
	settings.CloseDB()
}
