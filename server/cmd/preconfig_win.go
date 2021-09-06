//go:build windows
// +build windows

package main

import (
	"time"
)

const (
	EsSystemRequired = 0x00000001
	EsContinuous     = 0x80000000
)

var pulseTime = 1 * time.Minute

func Preconfig(kill bool) {
	// don't sleep/hibernate windows
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setThreadExecStateProc := kernel32.NewProc("SetThreadExecutionState")
	pulse := time.NewTicker(pulseTime)
	for {
		select {
		case <-pulse.C:
			setThreadExecStateProc.Call(uintptr(EsSystemRequired))
		}
	}
}
