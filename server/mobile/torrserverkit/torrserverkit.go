// Package torrserverkit exposes StartServer/StopServer/IsRunning for gomobile bind.
// This is a nested module so the gomobile tool dependency does not change server/go.mod.
package torrserverkit

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"server"
	"server/log"
	"server/settings"
)

var (
	runningMu sync.Mutex
	running   int32
)

// StartServer starts the embedded TorrServer HTTP engine and BitTorrent client.
// Returns an empty string on success, or an error message on failure.
func StartServer(port int, dataDir string) string {
	runningMu.Lock()
	defer runningMu.Unlock()

	if atomic.LoadInt32(&running) == 1 {
		return "server is already running"
	}

	if port <= 0 {
		port = 8090
	}
	portStr := strconv.Itoa(port)

	if dataDir == "" {
		dataDir, _ = os.Getwd()
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Sprintf("failed to create data directory: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:"+portStr)
	if err != nil {
		return fmt.Sprintf("port %s is already in use: %v", portStr, err)
	}
	_ = ln.Close()

	settings.Embedded = true
	settings.Path = dataDir
	log.Init(filepath.Join(dataDir, "torrserver.log"), filepath.Join(dataDir, "web.log"))

	settings.Args = &settings.ExecArgs{
		Port:     portStr,
		Path:     dataDir,
		RDB:      false,
		SearchWA: true,
	}

	if err := server.Start(); err != nil {
		return err.Error()
	}

	atomic.StoreInt32(&running, 1)
	go func() {
		defer atomic.StoreInt32(&running, 0)
		server.WaitServer()
	}()

	return ""
}

// StopServer stops the running TorrServer instance.
// Returns an empty string on success, or an error message on failure.
func StopServer() string {
	runningMu.Lock()
	defer runningMu.Unlock()

	if atomic.LoadInt32(&running) == 0 {
		return ""
	}

	done := make(chan struct{})
	go func() {
		server.Stop()
		close(done)
	}()

	select {
	case <-done:
		atomic.StoreInt32(&running, 0)
		return ""
	case <-time.After(5 * time.Second):
		atomic.StoreInt32(&running, 0)
		return "server shutdown timed out"
	}
}

// IsRunning returns true if TorrServer is actively running.
func IsRunning() bool {
	return atomic.LoadInt32(&running) == 1
}
