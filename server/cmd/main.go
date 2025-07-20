package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"github.com/alexflint/go-arg"
	"github.com/pkg/browser"

	"github.com/fsnotify/fsnotify"

	"server"
	"server/docs"
	"server/log"
	"server/settings"
	"server/torr"
	"server/version"
)

type args struct {
	Port        string `arg:"-p" help:"web server port (default 8090)"`
	IP          string `arg:"-i" help:"web server addr (default empty)"`
	Ssl         bool   `help:"enables https"`
	SslPort     string `help:"web server ssl port, If not set, will be set to default 8091 or taken from db(if stored previously). Accepted if --ssl enabled."`
	SslCert     string `help:"path to ssl cert file. If not set, will be taken from db(if stored previously) or default self-signed certificate/key will be generated. Accepted if --ssl enabled."`
	SslKey      string `help:"path to ssl key file. If not set, will be taken from db(if stored previously) or default self-signed certificate/key will be generated. Accepted if --ssl enabled."`
	Path        string `arg:"-d" help:"database and config dir path"`
	LogPath     string `arg:"-l" help:"server log file path"`
	WebLogPath  string `arg:"-w" help:"web access log file path"`
	RDB         bool   `arg:"-r" help:"start in read-only DB mode"`
	HttpAuth    bool   `arg:"-a" help:"enable http auth on all requests"`
	DontKill    bool   `arg:"-k" help:"don't kill server on signal"`
	UI          bool   `arg:"-u" help:"open torrserver page in browser"`
	TorrentsDir string `arg:"-t" help:"autoload torrents from dir"`
	TorrentAddr string `help:"Torrent client address, like 127.0.0.1:1337 (default :PeersListenPort)"`
	PubIPv4     string `arg:"-4" help:"set public IPv4 addr"`
	PubIPv6     string `arg:"-6" help:"set public IPv6 addr"`
	SearchWA    bool   `arg:"-s" help:"search without auth"`
	MaxSize     string `arg:"-m" help:"max allowed stream size (in Bytes)"`
	TGToken     string `arg:"-T" help:"telegram bot token"`
}

func (args) Version() string {
	return "TorrServer " + version.Version
}

var params args

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	arg.MustParse(&params)

	if params.Path == "" {
		params.Path, _ = os.Getwd()
	}

	if params.Port == "" {
		params.Port = "8090"
	}

	settings.Path = params.Path
	settings.HttpAuth = params.HttpAuth
	log.Init(params.LogPath, params.WebLogPath)
	fmt.Println("=========== START ===========")
	fmt.Println("TorrServer", version.Version+",", runtime.Version()+",", "CPU Num:", runtime.NumCPU())
	if params.HttpAuth {
		log.TLogln("Use HTTP Auth file", settings.Path+"/accs.db")
	}
	if params.RDB {
		log.TLogln("Running in Read-only DB mode!")
	}
	docs.SwaggerInfo.Version = version.Version

	dnsResolve()
	Preconfig(params.DontKill)

	if params.UI {
		go func() {
			time.Sleep(time.Second)
			if params.Ssl {
				browser.OpenURL("https://127.0.0.1:" + params.SslPort)
			} else {
				browser.OpenURL("http://127.0.0.1:" + params.Port)
			}
		}()
	}

	if params.TorrentAddr != "" {
		settings.TorAddr = params.TorrentAddr
	}

	if params.PubIPv4 != "" {
		settings.PubIPv4 = params.PubIPv4
	}

	if params.PubIPv6 != "" {
		settings.PubIPv6 = params.PubIPv6
	}

	if params.TorrentsDir != "" {
		go watchTDir(params.TorrentsDir)
	}

	if params.MaxSize != "" {
		maxSize, err := strconv.ParseInt(params.MaxSize, 10, 64)
		if err == nil {
			settings.MaxSize = maxSize
		}
	}

	server.Start(params.Port, params.IP, params.SslPort, params.SslCert, params.SslKey, params.Ssl, params.RDB, params.SearchWA, params.TGToken)
	log.TLogln(server.WaitServer())
	log.Close()
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

func dnsResolve() {
	addrs, err := net.LookupHost("www.google.com")
	if len(addrs) == 0 {
		log.TLogln("Check dns failed", addrs, err)

		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		}

		net.DefaultResolver = &net.Resolver{
			Dial: fn,
		}

		addrs, err = net.LookupHost("www.google.com")
		log.TLogln("Check cloudflare dns", addrs, err)
	} else {
		log.TLogln("Check dns OK", addrs, err)
	}
}

func watchTDir(dir string) {
	path, err := filepath.Abs(dir)
	if err != nil {
		path = dir
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.TLogln("Error creating watcher:", err)
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		log.TLogln("Error adding directory to watcher:", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				filename := event.Name
				if strings.ToLower(filepath.Ext(filename)) == ".torrent" {
					time.Sleep(2 * time.Second)
					sp, err := openFile(filename)
					if err != nil {
						log.TLogln("Error opening torrent file:", err)
						continue
					}
					tor, err := torr.AddTorrent(sp, "", "", "", "")
					if err != nil {
						log.TLogln("Error adding torrent:", err)
						continue
					}
					if tor.GotInfo() {
						if tor.Title == "" {
							tor.Title = tor.Name()
						}
						torr.SaveTorrentToDB(tor)
						tor.Drop()
						if err := os.Remove(filename); err != nil {
							log.TLogln("Error removing torrent file:", err)
						}
					} else {
						log.TLogln("Error getting torrent info")
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.TLogln("Watcher error:", err)
		}
	}
}

func openFile(path string) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	// mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	mag := minfo.Magnet(nil, &info)
	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: info.Name,
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}
