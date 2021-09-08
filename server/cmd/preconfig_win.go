//go:build windows
// +build windows

package main

import (
	"syscall"
	"time"

	"server/torr"
	"server/torr/state"
)

const (
	EsSystemRequired = 0x00000001
	EsContinuous     = 0x80000000
)

var pulseTime = 1 * time.Minute

func Preconfig(kill bool) {
	go func() {
		// don't sleep/hibernate windows
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		setThreadExecStateProc := kernel32.NewProc("SetThreadExecutionState")
		pulse := time.NewTicker(pulseTime)
		for {
			select {
			case <-pulse.C:
				{
					send := false
					for i, torrent := range torr.ListTorrent() {
						if torrent.Stat != state.TorrentInDB {
							send = true
							break
						}
					}
					if send {
						setThreadExecStateProc.Call(uintptr(EsSystemRequired))
					}
				}
			}
		}
	}()
}
