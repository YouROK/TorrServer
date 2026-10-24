package dlna

import (
	"runtime"
	"testing"
	"time"

	"server/leakcheck"
	"server/settings"
)

func TestStartStopNoGoroutineLeak(t *testing.T) {
	oldEmb, oldIPs, oldSets, oldNotify := settings.Embedded, settings.IPs, settings.BTsets, notifyInterval
	// Serve is stuck in Sleep(notifyInterval) until the period ends, then it
	// notices Close. The leak check only waits one second.
	notifyInterval = 20 * time.Millisecond
	settings.Embedded = true
	settings.IPs = []string{"127.0.0.1"}
	settings.BTsets = &settings.BTSets{FriendlyName: "DLNA Leak"}
	t.Cleanup(func() {
		Stop()
		settings.Embedded, settings.IPs, settings.BTsets = oldEmb, oldIPs, oldSets
		notifyInterval = oldNotify
	})

	base := runtime.NumGoroutine()
	Start()
	if dmsServer == nil {
		t.Fatal("dlna did not start")
	}
	Stop()
	if dmsServer != nil {
		t.Fatal("dlna still running")
	}
	leakcheck.AssertNoGoroutineLeak(t, base)
}
