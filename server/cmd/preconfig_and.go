//go:build android
// +build android

package main

// #cgo LDFLAGS: -static-libstdc++
import "C"

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"server"
	"server/log"
	"server/settings"
)

func Preconfig(dkill bool) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		for s := range sigc {
			if dkill {
				if settings.BTsets.EnableDebug || s != syscall.SIGPIPE {
					log.TLogln("Signal catched:", s)
					log.TLogln("To stop server, close it from web / api")
				}
				continue
			}

			log.TLogln("Signal catched:", s, "stopping server...")

			done := make(chan struct{})

			go func() {
				server.Stop()
				close(done)
			}()

			select {
			case <-done:
				log.TLogln("Server stopped gracefully")
			case <-time.After(5 * time.Second):
				log.TLogln("Server stop timeout, exiting forcefully")
				os.Exit(1)
			}
		}
	}()
}
