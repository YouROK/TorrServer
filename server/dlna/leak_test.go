package dlna

import (
	"runtime"
	"testing"

	"server/leakcheck"
	"server/settings"
)

func TestStartStopNoGoroutineLeak(t *testing.T) {
	oldEmb, oldIPs, oldSets := settings.Embedded, settings.IPs, settings.BTsets
	settings.Embedded = true
	settings.IPs = []string{"127.0.0.1"}
	settings.BTsets = &settings.BTSets{FriendlyName: "DLNA Leak"}
	t.Cleanup(func() {
		Stop()
		settings.Embedded, settings.IPs, settings.BTsets = oldEmb, oldIPs, oldSets
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
