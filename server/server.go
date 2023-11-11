package server

import (
	"net"
	"os"
	"path/filepath"

	"server/log"
	"server/settings"
	"server/web"
)

func Start(port, sslport, sslCert, sslKey string, sslEnabled, roSets, searchWA bool) {
	settings.InitSets(roSets, searchWA)

	//// https checks
	// check if ssl enabled
	settings.Ssl = sslEnabled
	settings.BTsets.Ssl = sslEnabled
	if sslEnabled {
		// set settings ssl enabled
		if sslport == "" {
			if settings.BTsets.SslPort == "" {
				settings.BTsets.SslPort = "8091"
			}
		} else {
			settings.BTsets.SslPort = sslport
		}
		// check if ssl cert and key files exist
		if sslCert != "" && sslKey != "" {
			// set settings ssl cert and key files
			settings.BTsets.SslCert = sslCert
			settings.BTsets.SslKey = sslKey
		}
		log.TLogln("Check web ssl port", settings.BTsets.SslPort)
		l, err := net.Listen("tcp", ":"+settings.BTsets.SslPort)
		if l != nil {
			l.Close()
		}
		if err != nil {
			log.TLogln("Port", settings.BTsets.SslPort, "already in use! Please set different port for HTTP. Abort")
			os.Exit(1)
		}
	}

	// http checks
	if port == "" {
		port = "8090"
	}
	log.TLogln("Check web port", port)
	l, err := net.Listen("tcp", ":"+port)
	if l != nil {
		l.Close()
	}
	if err != nil {
		log.TLogln("Port", port, "already in use! Please set different sslport for HTTPS. Abort")
		os.Exit(1)
	}

	// set settings http and https ports. Start web server.
	go cleanCache()
	settings.Port = port
	settings.SslPort = settings.BTsets.SslPort
	web.Start()

}

func cleanCache() {
	if !settings.BTsets.UseDisk || settings.BTsets.TorrentsSavePath == "/" || settings.BTsets.TorrentsSavePath == "" {
		return
	}

	dirs, err := os.ReadDir(settings.BTsets.TorrentsSavePath)
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
	files, err := os.ReadDir(path)
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
