package dlna

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/dms/dlna/dms"
	"github.com/wlynxg/anet"

	"server/netbind"
	"server/settings"
	"server/web/pages/template"
)

var dmsServer *dms.Server

func Start() {
	logger := slog.Default().With("component", "dlna")
	dmsServer = &dms.Server{
		Logger: logger.With("dms", "server"),
		Interfaces: func() (ifs []net.Interface) {
			var err error
			ifaces, err := anet.Interfaces()
			if err != nil {
				logger.Error("Failed to get network interfaces", "error", err)
				return
				// os.Exit(1) // avoid start on Android 13+
			}
			for _, i := range ifaces {
				// interface flags seem to always be 0 on Windows
				if runtime.GOOS != "windows" && (i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 || i.Flags&net.FlagMulticast == 0) {
					continue
				}
				ifs = append(ifs, i)
			}
			return
		}(),
		HTTPConn: func() net.Listener {
			port := 9080
			for {
				logger.Info("Checking dlna port", "port", port)
				if err := netbind.CheckPort(settings.IPs, strconv.Itoa(port)); err == nil {
					break
				}
				port++
			}
			logger.Info("Setting dlna port", "port", port)
			conn, err := netbind.Listen(settings.IPs, strconv.Itoa(port))
			if err != nil {
				logger.Error("Failed to listen on bind port", "error", err)
				os.Exit(1)
			}
			return conn
		}(),
		FriendlyName:        getDefaultFriendlyName(),
		NoTranscode:         true,
		NoProbe:             true,
		StallEventSubscribe: false,
		Icons: []dms.Icon{
			{
				Width:    48,
				Height:   48,
				Depth:    24,
				Mimetype: "image/png",
				Bytes:    template.Dlnaicon48png,
			},
			{
				Width:    120,
				Height:   120,
				Depth:    24,
				Mimetype: "image/png",
				Bytes:    template.Dlnaicon120png,
			},
		},
		LogHeaders:     settings.BTsets.EnableDebug,
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
		logger.Error("Error initializing dms server", "error", err)
		os.Exit(1)
	}
	go func() {
		if err := dmsServer.Run(); err != nil {
			logger.Error("DMS server run crash", "error", err)
			os.Exit(1)
		}
	}()
}

func Stop() {
	if dmsServer != nil {
		dmsServer.Close()
		dmsServer = nil
	}
}

func onBrowse(path, rootObjectPath, host, userAgent string) (ret []interface{}, err error) {
	if path == "/" {
		ret = getRoot()
		return
	} else if path == "/TR" {
		ret = getTorrentCategories()
		return
	} else if strings.HasPrefix(path, "/TR/") {
		ret = getTorrentsByCategory(path)
		return
	} else if isHashPath(path) {
		ret = getTorrent(path, host)
		return
	} else if filepath.Base(path) == "LD" {
		ret = loadTorrent(path, host)
	}
	return
}

func onBrowseMeta(path string, rootObjectPath string, host, userAgent string) (ret interface{}, err error) {
	ret = getTorrentMeta(path, host)
	if ret == nil {
		err = fmt.Errorf("meta not found")
	}
	return
}

func getDefaultFriendlyName() string {
	logger := slog.Default().With("component", "dlna")

	if settings.BTsets.FriendlyName != "" {
		return settings.BTsets.FriendlyName
	}

	ret := "TorrServer"
	userName := ""
	user, err := user.Current()
	if err != nil {
		logger.Warn("getDefaultFriendlyName could not get username", "error", err)
	} else {
		userName = user.Name
	}
	host, err := os.Hostname()
	if err != nil {
		logger.Warn("getDefaultFriendlyName could not get hostname", "error", err)
	}

	if userName == "" && host == "" {
		return ret
	}

	if userName != "" && host != "" {
		if userName == host {
			return ret + ": " + userName
		}
		return ret + ": " + userName + " on " + host
	}

	if host == "localhost" { // useless host, use 1st IP
		ifaces, err := anet.Interfaces()
		if err != nil {
			return ret + ": " + userName + "@" + host
		}
		var list []string
		for _, i := range ifaces {
			// interface flags seem to always be 0 on Windows
			if runtime.GOOS != "windows" && (i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 || i.Flags&net.FlagMulticast == 0) {
				continue
			}
			addrs, _ := anet.InterfaceAddrsByInterface(&i)
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if !ip.IsLoopback() && ip.To4() != nil {
					list = append(list, ip.String())
				}
			}
		}
		if len(list) > 0 {
			sort.Strings(list)
			return ret + " " + list[0]
		}
	}
	return ret + ": " + userName + "@" + host
}
