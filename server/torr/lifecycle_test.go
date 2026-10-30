package torr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/leakcheck"
	"server/settings"
	"server/torr/offlinetest"
)

func TestAddExtraTrackersKeepsUnion(t *testing.T) {
	spec := &torrent.TorrentSpec{Trackers: [][]string{{"udp://original.example:1/announce"}}}
	var remote, local []string
	for i := 1; i <= 20; i++ {
		remote = append(remote, fmt.Sprintf("udp://198.51.100.%d:1337/announce", i))
	}
	for i := 1; i <= 14; i++ {
		local = append(local, fmt.Sprintf("udp://local%d.example:80/announce", i))
	}
	local = append(local, "https://1.2.3.4/announce", "udp://198.51.100.1:1337/announce")
	addExtraTrackers(spec, local, remote)
	if len(spec.Trackers) != 2 {
		t.Fatalf("tiers = %d, want original + one extra", len(spec.Trackers))
	}
	extra := spec.Trackers[1]
	if len(extra) != 34 {
		t.Fatalf("extra len = %d, want 34", len(extra))
	}
	seen := map[string]struct{}{}
	for _, tr := range extra {
		if strings.HasPrefix(tr, "https://") {
			t.Fatalf("raw-ip https kept: %s", tr)
		}
		seen[tr] = struct{}{}
	}
	for i := 1; i <= 14; i++ {
		want := fmt.Sprintf("udp://local%d.example:80/announce", i)
		if _, ok := seen[want]; !ok {
			t.Fatalf("dropped %s", want)
		}
	}
	for i := 1; i <= 20; i++ {
		want := fmt.Sprintf("udp4://198.51.100.%d:1337/announce", i)
		if _, ok := seen[want]; !ok {
			t.Fatalf("dropped %s", want)
		}
	}
}

func TestResolveDHTBootstrapSkipsIPv6(t *testing.T) {
	old := lookupDHTHost
	t.Cleanup(func() { lookupDHTHost = old })
	var looked []string
	lookupDHTHost = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		looked = append(looked, host)
		if ip := net.ParseIP(host); ip != nil {
			return []net.IPAddr{{IP: ip}}, nil
		}
		return nil, context.DeadlineExceeded
	}
	addrs, err := resolveDHTBootstrap(context.Background(), []string{
		"1.2.3.4:6881",
		"[2001:db8::1]:6881",
		"router.silotis.us:6881",
	}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || !addrs[0].IP().Equal(net.ParseIP("1.2.3.4")) {
		t.Fatalf("addrs = %v", addrs)
	}
	for _, host := range looked {
		if strings.EqualFold(host, "router.silotis.us") {
			t.Fatal("looked up ipv6-only bootstrap")
		}
	}
}

func TestResolveDHTBootstrapTimeout(t *testing.T) {
	old := lookupDHTHost
	t.Cleanup(func() { lookupDHTHost = old })
	lookupDHTHost = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := resolveDHTBootstrap(ctx, []string{"router.bittorrent.com:6881", "dht.transmissionbt.com:6881"}, true, false)
	if err == nil {
		t.Fatal("expected error")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("dns took %s", elapsed)
	}
}

func TestBTServerDHTLifecycle(t *testing.T) {
	offlinetest.Prepare(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("udp://127.0.0.1:9/announce\n"))
	}))
	t.Cleanup(srv.Close)
	settings.BTsets.TrackersListURL = srv.URL
	settings.BTsets.DefaultTrackers = "udp://127.0.0.1:9/announce"

	oldHosts, oldLookup := dhtBootstrapHosts, lookupDHTHost
	t.Cleanup(func() {
		dhtBootstrapHosts = oldHosts
		lookupDHTHost = oldLookup
	})

	settings.BTsets.DisableDHT = true
	bt := NewBTS()
	baseG := runtime.NumGoroutine()
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	bt.Disconnect()
	leakcheck.AssertNoGoroutineLeak(t, baseG)

	lookupDHTHost = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	dhtBootstrapHosts = []string{"router.bittorrent.com:6881", "router.silotis.us:6881"}
	settings.BTsets.DisableDHT = false
	settings.BTsets.EnableIPv6 = false

	baseG = runtime.NumGoroutine()
	start := time.Now()
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("Connect blocked for %s", time.Since(start))
	}
	if len(bt.client.DhtServers()) == 0 {
		t.Fatal("DHT enabled but no servers")
	}
	time.Sleep(50 * time.Millisecond)
	start = time.Now()
	bt.Disconnect()
	if time.Since(start) > 2*time.Second {
		t.Fatalf("Disconnect blocked for %s", time.Since(start))
	}
	leakcheck.AssertNoGoroutineLeak(t, baseG)

	// Second cycle: Connect shuts the previous client down, then starts another.
	baseG = runtime.NumGoroutine()
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	bt.Disconnect()
	leakcheck.AssertNoGoroutineLeak(t, baseG)
}

// TestClientCloseStopsDHTServe closes only through torrent.Client.Close.
// BTServer.Disconnect also closes DHT servers before that, so this is the check
// that the client fork itself wakes the uTP read.
func TestClientCloseStopsDHTServe(t *testing.T) {
	offlinetest.Prepare(t)
	oldHosts, oldLookup := dhtBootstrapHosts, lookupDHTHost
	t.Cleanup(func() {
		dhtBootstrapHosts = oldHosts
		lookupDHTHost = oldLookup
	})
	lookupDHTHost = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("192.0.2.1")}}, nil
	}
	dhtBootstrapHosts = []string{"router.bittorrent.com:6881"}
	settings.BTsets.DisableDHT = false
	settings.BTsets.EnableIPv6 = false

	bt := NewBTS()
	var cancel context.CancelFunc
	bt.dhtCtx, cancel = context.WithCancel(context.Background())
	t.Cleanup(cancel)
	bt.configure()

	baseG := runtime.NumGoroutine()
	client, err := torrent.NewClient(bt.config)
	if err != nil {
		t.Fatal(err)
	}
	if len(client.DhtServers()) == 0 {
		client.Close()
		t.Fatal("DHT enabled but no servers")
	}
	time.Sleep(50 * time.Millisecond)
	client.Close()
	leakcheck.AssertNoGoroutineLeak(t, baseG)
}

func TestAddExtraTrackersPinsLiteralIPFamily(t *testing.T) {
	magnet := []string{"udp://198.51.100.1:1337/announce", "udp://open.example:80/announce"}
	spec := &torrent.TorrentSpec{Trackers: [][]string{magnet}}
	addExtraTrackers(spec, []string{
		"udp://198.51.100.1:1337/announce",
		"udp://[2001:db8::1]:6969/announce",
		"udp://other.example:80/announce",
	})
	if len(spec.Trackers) != 2 {
		t.Fatalf("tiers = %d, want magnet tier + one extra", len(spec.Trackers))
	}
	wantMagnet := []string{"udp4://198.51.100.1:1337/announce", "udp://open.example:80/announce"}
	if !reflect.DeepEqual(spec.Trackers[0], wantMagnet) {
		t.Fatalf("magnet tier = %v, want %v", spec.Trackers[0], wantMagnet)
	}
	wantExtra := []string{
		"udp4://198.51.100.1:1337/announce",
		"udp6://[2001:db8::1]:6969/announce",
		"udp://other.example:80/announce",
	}
	if !reflect.DeepEqual(spec.Trackers[1], wantExtra) {
		t.Fatalf("extra tier = %v, want %v", spec.Trackers[1], wantExtra)
	}
	if magnet[0] != "udp://198.51.100.1:1337/announce" {
		t.Fatalf("caller's tier modified: %v", magnet)
	}
}
