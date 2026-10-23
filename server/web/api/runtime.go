package api

import (
	"net/http"
	"runtime"

	sets "server/settings"
	"server/torr"

	"github.com/gin-gonic/gin"
)

// RuntimeMemory is a detailed process + torrstor-cache snapshot for leak/OOM triage.
type RuntimeMemory struct {
	NumGoroutine int `json:"num_goroutine"`
	GOMAXPROCS   int `json:"gomaxprocs"`

	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapReleased uint64 `json:"heap_released"`
	HeapObjects  uint64 `json:"heap_objects"`

	StackInuse uint64 `json:"stack_inuse"`
	StackSys   uint64 `json:"stack_sys"`
	Sys        uint64 `json:"sys"`

	TotalAlloc  uint64 `json:"total_alloc"`
	Mallocs     uint64 `json:"mallocs"`
	Frees       uint64 `json:"frees"`
	LiveObjects uint64 `json:"live_objects"`

	NextGC        uint64  `json:"next_gc"`
	LastGC        uint64  `json:"last_gc_unix_ns"`
	NumGC         uint32  `json:"num_gc"`
	NumForcedGC   uint32  `json:"num_forced_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
	PauseTotalNs  uint64  `json:"gc_pause_total_ns"`
	PauseNsLast   uint64  `json:"gc_pause_ns_last"`

	CacheFilled   int64 `json:"cache_filled"`
	CacheCapacity int64 `json:"cache_capacity"`
	CacheCount    int   `json:"cache_count"`
}

// RuntimeStatus is a read-only snapshot for the web UI: integrations + BT + memory.
type RuntimeStatus struct {
	DLNAEnabled    bool   `json:"dlna_enabled"`
	BonjourEnabled bool   `json:"bonjour_enabled"`
	FriendlyName   string `json:"friendly_name"`
	WebDAVEnabled  bool   `json:"webdav_enabled"`
	WebDAVPath     string `json:"webdav_path"`
	FusePath       string `json:"fuse_path"`
	FuseEnabled    bool   `json:"fuse_enabled"`

	Memory *RuntimeMemory  `json:"memory"`
	Caches []torr.CacheRow `json:"caches,omitempty"`

	// BitTorrent client (structured alternative to plaintext GET /stat).
	BT *torr.ClientStatusSnapshot `json:"bt,omitempty"`
}

func snapshotRuntimeMemory(cacheFilled, cacheCapacity int64, cacheCount int) *RuntimeMemory {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	pauseLast := uint64(0)
	if m.NumGC > 0 {
		pauseLast = m.PauseNs[(m.NumGC+255)%256]
	}
	return &RuntimeMemory{
		NumGoroutine:  runtime.NumGoroutine(),
		GOMAXPROCS:    runtime.GOMAXPROCS(0),
		HeapAlloc:     m.HeapAlloc,
		HeapInuse:     m.HeapInuse,
		HeapIdle:      m.HeapIdle,
		HeapSys:       m.HeapSys,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		Sys:           m.Sys,
		TotalAlloc:    m.TotalAlloc,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		LiveObjects:   m.Mallocs - m.Frees,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		PauseTotalNs:  m.PauseTotalNs,
		PauseNsLast:   pauseLast,
		CacheFilled:   cacheFilled,
		CacheCapacity: cacheCapacity,
		CacheCount:    cacheCount,
	}
}

// runtimeStatus godoc
//
//	@Summary		Runtime integration + BitTorrent + memory status
//	@Description	Read-only flags for DLNA/Bonjour/WebDAV/FUSE, MemStats + per-torrent cache filled, and structured BT stats.
//	@Tags			API
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	RuntimeStatus
//	@Router			/runtime/status [get]
func runtimeStatus(c *gin.Context) {
	friendly := ""
	dlna := false
	bonjour := false
	if sets.BTsets != nil {
		friendly = sets.BTsets.FriendlyName
		dlna = sets.BTsets.EnableDLNA
		bonjour = sets.BTsets.EnableBonjour
	}
	fusePath := ""
	webdav := false
	if sets.Args != nil {
		fusePath = sets.Args.FusePath
		webdav = sets.Args.WebDAV
	}
	caches, filled, capacity := torr.SnapshotCaches()
	bt := torr.SnapshotClientStatus()
	c.JSON(http.StatusOK, RuntimeStatus{
		DLNAEnabled:    dlna,
		BonjourEnabled: bonjour,
		FriendlyName:   friendly,
		WebDAVEnabled:  webdav,
		WebDAVPath:     "/dav",
		FusePath:       fusePath,
		FuseEnabled:    fusePath != "",
		Memory:         snapshotRuntimeMemory(filled, capacity, len(caches)),
		Caches:         caches,
		BT:             &bt,
	})
}
