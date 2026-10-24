package main

import (
	"sync"
	"testing"
	"time"
)

func TestIsSuspiciousAddress(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"", false},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"10.1.2.3", false}, // private 10/8 is intentionally not treated as hijacking
		{"127.0.0.1", true},
		{"127.0.0.10", true},
		{"0.0.0.0", true},
		{"::1", true},
		{"192.168.1.1", true},
		{"169.254.1.1", true},
		{"2001:4860:4860::8888", false},
	}
	for _, tc := range cases {
		if got := isSuspiciousAddress(tc.addr); got != tc.want {
			t.Errorf("isSuspiciousAddress(%q) = %v, want %v", tc.addr, got, tc.want)
		}
	}
}

func TestNewDNSCheckerDefaults(t *testing.T) {
	d := NewDNSChecker(DNSConfig{})
	if len(d.config.PrimaryServers) != len(DefaultDNSConfig().PrimaryServers) {
		t.Fatalf("primary = %#v", d.config.PrimaryServers)
	}
	if d.config.Timeout != 5*time.Second || d.config.CacheDuration != 5*time.Minute {
		t.Fatalf("timeouts = %s %s", d.config.Timeout, d.config.CacheDuration)
	}
	if d.cache == nil || d.cacheTime == nil {
		t.Fatal("cache maps not initialized")
	}
}

func TestDNSCacheHitMissAndExpiry(t *testing.T) {
	d := NewDNSChecker(DNSConfig{
		PrimaryServers: []string{"127.0.0.1:1"},
		Timeout:        time.Millisecond,
		CacheDuration:  time.Hour,
	})
	if _, ok := d.getFromCache("missing.test"); ok {
		t.Fatal("missing key hit")
	}

	in := []string{"1.2.3.4", "5.6.7.8"}
	d.addToCache("example.test", in)
	in[0] = "mutated"
	got, err := d.LookupHostWithFallback("example.test")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "1.2.3.4" || got[1] != "5.6.7.8" {
		t.Fatalf("cache = %#v", got)
	}
	got[0] = "caller"
	again, _ := d.getFromCache("example.test")
	if again[0] != "1.2.3.4" {
		t.Fatalf("caller mutated cache: %#v", again)
	}

	d.config.CacheDuration = 0
	if _, ok := d.getFromCache("example.test"); ok {
		t.Fatal("expired entry returned")
	}
	if _, ok := d.getFromCache("example.test"); ok {
		t.Fatal("expired entry still present")
	}
}

func TestDNSCacheConcurrent(t *testing.T) {
	for _, ttl := range []time.Duration{time.Hour, 0} {
		d := NewDNSChecker(DNSConfig{
			PrimaryServers: []string{"127.0.0.1:1"},
			Timeout:        time.Millisecond,
			CacheDuration:  ttl,
		})
		var wg sync.WaitGroup
		for g := 0; g < 40; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 80; i++ {
					d.addToCache("h", []string{"1.1.1.1", "8.8.8.8"})
					if addrs, ok := d.getFromCache("h"); ok && len(addrs) != 2 {
						t.Errorf("cache hit = %#v", addrs)
					}
				}
			}()
		}
		wg.Wait()
	}
}
