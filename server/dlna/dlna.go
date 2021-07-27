package dlna

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/anacrolix/dms/dlna/dms"

	"server/log"
	"server/web/pages/template"
)

var dmsServer *dms.Server

func Start() {
	dmsServer = &dms.Server{
		Interfaces: func() (ifs []net.Interface) {
			var err error
			ifs, err = net.Interfaces()
			if err != nil {
				log.TLogln(err)
				os.Exit(1)
			}
			return
		}(),
		HTTPConn: func() net.Listener {
			conn, err := net.Listen("tcp", ":9080")
			if err != nil {
				log.TLogln(err)
				os.Exit(1)
			}
			return conn
		}(),
		FriendlyName: "TorrServer",
		NoTranscode:  true,
		NoProbe:      true,
		Icons: []dms.Icon{
			dms.Icon{
				Width:      192,
				Height:     192,
				Depth:      32,
				Mimetype:   "image/png",
				ReadSeeker: bytes.NewReader(template.Androidchrome192x192png),
			},
			dms.Icon{
				Width:      32,
				Height:     32,
				Depth:      32,
				Mimetype:   "image/png",
				ReadSeeker: bytes.NewReader(template.Favicon32x32png),
			},
		},
		NotifyInterval: 30 * time.Second,
		AllowedIpNets: func() []*net.IPNet {
			var nets []*net.IPNet
			_, ipnet, _ := net.ParseCIDR("0.0.0.0/0")
			nets = append(nets, ipnet)
			_, ipnet, _ = net.ParseCIDR("::/0")
			nets = append(nets, ipnet)
			return nets
		}(),
		OnBrowseDirectChildren: onBrowse,
		OnBrowseMetadata:       onBrowseMeta,
	}

	if err := dmsServer.Init(); err != nil {
		log.TLogln("error initing dms server: %v", err)
		os.Exit(1)
	}
	go func() {
		if err := dmsServer.Run(); err != nil {
			log.TLogln(err)
			os.Exit(1)
		}
	}()
}

func Stop() {
	if dmsServer != nil {
		dmsServer.Close()
	}
}

func onBrowse(path, rootObjectPath, host, userAgent string) (ret []interface{}, err error) {
	if path == "/" {
		ret = getTorrents()
		return
	} else if isHashPath(path) {
		ret = getTorrent(path, host)
		return
	} else if filepath.Base(path) == "Load Torrent" {
		ret = loadTorrent(path, host)
	}
	return
}

func onBrowseMeta(path string, rootObjectPath string, host, userAgent string) (ret interface{}, err error) {
	err = fmt.Errorf("not implemented")
	return
}
