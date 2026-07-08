package main

import (
	"testing"

	"github.com/alexflint/go-arg"
)

func TestIPsSeparateFlag(t *testing.T) {
	var params args
	p, err := arg.NewParser(arg.Config{}, &params)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--ip", "127.0.0.1", "--ip", "192.168.1.100"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"127.0.0.1", "192.168.1.100"}
	if len(params.IPs) != len(want) {
		t.Fatalf("IPs = %#v, want %#v", params.IPs, want)
	}
	for i := range want {
		if params.IPs[i] != want[i] {
			t.Fatalf("IPs = %#v, want %#v", params.IPs, want)
		}
	}
}
