package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"server"
	"server/log"
	"server/settings"
	"server/version"
)

type args struct {
	Port     string `arg:"-p" help:"web server port"`
	Path     string `arg:"-d" help:"database path"`
	RDB      bool   `arg:"-r" help:"start in read-only DB mode"`
	DontKill bool   `arg:"-k" help:"dont kill program on signal"`
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
	dnsResolve()
	Preconfig(params.DontKill)

	server.Start(params.Port, params.RDB)
	log.TLogln(server.WaitServer())
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

func dnsResolve() {
	addrs, err := net.LookupHost("www.themoviedb.org")
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
