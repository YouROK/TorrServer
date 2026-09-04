package api

import (
	"net/http"

	sets "server/settings"
	"server/torr"

	"github.com/gin-gonic/gin"
)

// RuntimeStatus is a read-only snapshot for the web UI: integrations + BT client stats.
type RuntimeStatus struct {
	DLNAEnabled    bool   `json:"dlna_enabled"`
	BonjourEnabled bool   `json:"bonjour_enabled"`
	FriendlyName   string `json:"friendly_name"`
	WebDAVEnabled  bool   `json:"webdav_enabled"`
	WebDAVPath     string `json:"webdav_path"`
	FusePath       string `json:"fuse_path"`
	FuseEnabled    bool   `json:"fuse_enabled"`

	// BitTorrent client (structured alternative to plaintext GET /stat).
	BT *torr.ClientStatusSnapshot `json:"bt,omitempty"`
}

// runtimeStatus godoc
//
//	@Summary		Runtime integration + BitTorrent status
//	@Description	Read-only flags for DLNA/Bonjour/WebDAV/FUSE plus structured BT stats (and raw /stat text).
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
	fusePath := sets.Args.FusePath
	bt := torr.SnapshotClientStatus()
	c.JSON(http.StatusOK, RuntimeStatus{
		DLNAEnabled:    dlna,
		BonjourEnabled: bonjour,
		FriendlyName:   friendly,
		WebDAVEnabled:  sets.Args.WebDAV,
		WebDAVPath:     "/dav",
		FusePath:       fusePath,
		FuseEnabled:    fusePath != "",
		BT:             &bt,
	})
}
