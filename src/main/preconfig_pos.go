// +build !windows

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func Preconfig(kill bool) {
	if kill {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGSTOP,
			syscall.SIGPIPE,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		go func() {
			for s := range sigc {
				fmt.Println("Signal catched:", s)
				fmt.Println("For stop server, close in web")
			}
		}()
	}

	//dns resover
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
