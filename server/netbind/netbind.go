package netbind

import (
	"fmt"
	"net"
	"sync"
)

// Normalize returns unique bind hosts. An empty list means all interfaces once.
func Normalize(ips []string) []string {
	if len(ips) == 0 {
		return []string{""}
	}
	out := make([]string, 0, len(ips))
	seen := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		out = append(out, ip)
	}
	return out
}

// Addr formats host:port for net.Listen.
func Addr(host, port string) string {
	if host == "" {
		return ":" + port
	}
	return net.JoinHostPort(host, port)
}

// CheckPort verifies that TCP port is available on every bind host.
func CheckPort(hosts []string, port string) error {
	for _, host := range Normalize(hosts) {
		addr := Addr(host, port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("%s: %w", addr, err)
		}
		l.Close()
	}
	return nil
}

// Listen binds TCP on each host and returns a listener that accepts from all of them.
func Listen(hosts []string, port string) (net.Listener, error) {
	hosts = Normalize(hosts)
	if len(hosts) == 1 {
		return net.Listen("tcp", Addr(hosts[0], port))
	}

	ml := &multiListener{
		conns: make(chan net.Conn),
		errs:  make(chan error, len(hosts)),
	}
	for _, host := range hosts {
		l, err := net.Listen("tcp", Addr(host, port))
		if err != nil {
			ml.Close()
			return nil, err
		}
		ml.listeners = append(ml.listeners, l)
		go ml.acceptLoop(l)
	}
	return ml, nil
}

type multiListener struct {
	listeners []net.Listener
	conns     chan net.Conn
	errs      chan error
	closeOnce sync.Once
}

func (ml *multiListener) acceptLoop(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			select {
			case ml.errs <- err:
			default:
			}
			return
		}
		ml.conns <- c
	}
}

func (ml *multiListener) Accept() (net.Conn, error) {
	select {
	case c := <-ml.conns:
		return c, nil
	case err := <-ml.errs:
		return nil, err
	}
}

func (ml *multiListener) Close() error {
	var err error
	ml.closeOnce.Do(func() {
		for _, l := range ml.listeners {
			if closeErr := l.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}
	})
	return err
}

func (ml *multiListener) Addr() net.Addr {
	if len(ml.listeners) == 0 {
		return nil
	}
	return ml.listeners[0].Addr()
}
