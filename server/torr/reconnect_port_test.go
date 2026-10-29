package torr

import (
	"net"
	"strconv"
	"testing"

	"github.com/anacrolix/utp"

	"server/settings"
	"server/torr/offlinetest"
)

// Settings save is Disconnect then Connect on the same PeersListenPort. A peer
// still connected over uTP used to hold the port for ~20s ("address already in use").
func TestReconnectSamePortWithLiveUtpPeer(t *testing.T) {
	offlinetest.Prepare(t)
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := pc.LocalAddr().(*net.UDPAddr).Port
	_ = pc.Close()
	settings.BTsets.DisableUTP = false
	settings.TorAddr = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	bt := NewBTS()
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(bt.Disconnect)

	peer, err := utp.NewSocket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = peer.Close() }()
	conn, err := peer.Dial(net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	bt.Disconnect()
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect on port %d right after Disconnect: %v", port, err)
	}
}
