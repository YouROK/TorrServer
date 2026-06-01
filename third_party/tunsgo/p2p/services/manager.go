package services

import (
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type P2PService interface {
	Start() error
	Stop()
	Name() string
	ProtocolID() protocol.ID
	HandleStream(network.Stream)
}

type Manager struct {
	Host     host.Host
	Services []P2PService
}

func NewManager(h host.Host) *Manager {
	return &Manager{
		Host:     h,
		Services: make([]P2PService, 0),
	}
}

func (m *Manager) GetServices() []P2PService {
	return m.Services
}

func (m *Manager) AddService(srv P2PService) {
	m.Services = append(m.Services, srv)
	if srv.ProtocolID() != "" {
		m.Host.SetStreamHandler(srv.ProtocolID(), srv.HandleStream)
	}
}

func (m *Manager) Start() error {
	for _, srv := range m.Services {
		err := srv.Start()
		if err != nil {
			log.Println("[P2P Srvc] Error start service:", srv.Name(), err)
			return err
		}
	}
	return nil
}

func (m *Manager) Stop() {
	for _, srv := range m.Services {
		srv.Stop()
	}
}
