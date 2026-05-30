package models

import "time"

type PeerInfo struct {
	PeerID    string    `json:"peer_id"`
	Hosts     []string  `json:"hosts,omitempty"`
	Timestamp int64     `json:"timestamp"`
	LastResp  time.Time `json:"-"`
	LastSeen  time.Time `json:"-"`
}
