package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/pkg/browser"

	"server"
	"server/log"
	"server/settings"
	"server/torr"
	utils2 "server/utils"
	"server/version"
	"server/web/api/utils"
)

type args struct {
	Port     string `arg:"-p" help:"web server port"`
	Path     string `arg:"-d" help:"database path"`
	LogPath  string `arg:"-l" help:"log path"`
	RDB      bool   `arg:"-r" help:"start in read-only DB mode"`
	HttpAuth bool   `arg:"-a" help:"http auth on all requests"`
	DontKill bool   `arg:"-k" help:"dont kill server on signal"`
	UI       bool   `arg:"-u" help:"run page torrserver in browser"`
}

func (args) Version() string {
	return "TorrServer " + version.Version
}

var params args

func main() {
	arg.MustParse(&params)

	if params.Path == "" {
		params.Path, _ = os.Getwd()
	}

	if params.Port == "" {
		params.Port = "8090"
	}

	settings.Path = params.Path
	settings.HttpAuth = params.HttpAuth
	log.Init(params.LogPath)

	dnsResolve()

	if params.UI {
		go func() {
			time.Sleep(time.Second)
			browser.OpenURL("http://127.0.0.1:" + params.Port)
		}()
	}

	go testLoad()
	server.Start(params.Port, params.RDB)
	log.TLogln(server.WaitServer())
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

func dnsResolve() {
	addrs, err := net.LookupHost("www.google.com")
	if len(addrs) == 0 {
		fmt.Println("Check dns", addrs, err)

		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		}

		net.DefaultResolver = &net.Resolver{
			Dial: fn,
		}

		addrs, err = net.LookupHost("www.themoviedb.org")
		fmt.Println("Check new dns", addrs, err)
	}
}

func testLoad() {
	time.Sleep(time.Second * 5)
	spec, err := utils.ParseLink("magnet:?xt=urn:btih:41951b387b5e42cb72be44df14040c5d138770bb")
	if err != nil {
		fmt.Println(err)
		return
	}

	tor, err := torr.AddTorrent(spec, "Test", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	tor.GotInfo()
	files := tor.Files()
	if len(files) == 0 {
		fmt.Println("Files is empty")
		return
	}

	torr.Preload(tor, 0)

	buf := make([]byte, 32*1024, 32*1024)
	offset := 0
	readed := 0

	lastTime := time.Now()
	fmt.Println("Read file...")
	reader := tor.NewReader(files[0])
	for {
		n, err := reader.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		offset += n
		readed += n
		since := time.Since(lastTime)
		if since > time.Second {
			readStr := utils2.Format(float64(readed) / since.Seconds())
			loadStr := utils2.Format(tor.Status().DownloadSpeed)
			fmt.Println("RS:", readStr+"/sec", "DS:", loadStr+"/sec")
			readed = 0
			lastTime = time.Now()

		}
	}
}
