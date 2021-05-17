// +build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Preconfig(dkill bool) {
	if dkill {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,

			syscall.SIGPIPE,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		go func() {
			for s := range sigc {
				if dkill {
					fmt.Println("Signal catched:", s)
					fmt.Println("For stop server, close in api")
				}
			}
		}()
	}
}
