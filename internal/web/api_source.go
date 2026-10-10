package web

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"silo/internal/log"
	"silo/internal/torrent"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// sourceOpener открывает файл раздачи для внутреннего источника.
type sourceOpener func(u *user.User, hash string, fileIdx int) (io.ReadSeekCloser, *torrent.TorrentFileStat, error)

// openSource возвращает файл раздачи, используя подменяемый открыватель.
func (s *Server) openSource(u *user.User, hash string, fileIdx int) (io.ReadSeekCloser, *torrent.TorrentFileStat, error) {
	if s.sourceOpen != nil {
		return s.sourceOpen(u, hash, fileIdx)
	}
	return s.torrentMgr.GetStreamReader(u, hash, fileIdx)
}

// SetSourceOpener подменяет способ открытия файла раздачи.
func (s *Server) SetSourceOpener(fn sourceOpener) {
	s.sourceOpen = fn
}

// handleInternalSource отдаёт файл раздачи процессу ffmpeg по временному доступу.
// Роут не защищён токеном: ffmpeg не умеет передавать заголовки, а токен в URL
// попал бы в список процессов. Доступ ограничен адресом отправителя и тем,
// что идентификатор доступа непредсказуем и живёт ограниченное время.
func (s *Server) handleInternalSource(c *gin.Context) {
	if !s.isLocalRequest(c) {
		log.Warnf("[Web] Rejected internal source request from %s", c.RemoteIP())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	lease, ok := s.sourceRegistry.Get(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Транскодирование длинного файла занимает больше времени, чем срок доступа,
	// поэтому каждый запрос продлевает его
	s.sourceRegistry.Touch(lease.ID)

	u, err := s.userSvc.GetUserByID(lease.UserID)
	if err != nil {
		log.Errorf("[Web] Source lease %s refers to unknown user %s: %v", lease.ID, lease.UserID, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	reader, fileStat, err := s.openSource(u, lease.Hash, lease.FileIdx)
	if err != nil {
		log.Errorf("[Web] Failed to open source %s/%d: %v", lease.Hash, lease.FileIdx, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer func() { _ = reader.Close() }()

	// Range-запросы обязательны: ffmpeg читает mp4 с moov в конце через перемотку
	http.ServeContent(c.Writer, c.Request, fileStat.Path, s.startedAt, reader)
}

// isLocalRequest проверяет, что запрос пришёл с этой же машины.
func (s *Server) isLocalRequest(c *gin.Context) bool {
	ip := net.ParseIP(c.RemoteIP())
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}

	// Если сервер привязан к конкретному адресу, loopback недоступен и ffmpeg
	// обращается к нему по адресу привязки, поэтому его тоже считаем локальным.
	bind := s.bindIP()
	return bind != nil && bind.Equal(ip)
}

// bindIP возвращает адрес привязки сервера, если он задан явно.
func (s *Server) bindIP() net.IP {
	host := strings.Trim(strings.TrimSpace(s.cfg.Server.Host), "[]")
	if host == "" || host == "0.0.0.0" || host == "::" {
		return nil
	}
	return net.ParseIP(host)
}

// SourceURL собирает адрес файла для ffmpeg по выданному доступу.
func (s *Server) SourceURL(id string) string {
	return s.sourceBaseURL() + "/api/internal/source/" + id
}

// sourceBaseURL возвращает базовый адрес сервера, доступный с этой же машины.
func (s *Server) sourceBaseURL() string {
	s.sourceMu.RLock()
	host := s.sourceHost
	port := s.sourcePort
	s.sourceMu.RUnlock()

	if host == "" {
		host = s.defaultSourceHost()
	}
	if port == 0 {
		port = s.cfg.Server.Port
	}
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}

// defaultSourceHost подбирает адрес, по которому сервер доступен локально.
func (s *Server) defaultSourceHost() string {
	if ip := s.bindIP(); ip != nil {
		return ip.String()
	}
	return "127.0.0.1"
}

// setSourceAddr запоминает фактический адрес прослушивания.
func (s *Server) setSourceAddr(addr net.Addr) {
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		host = s.defaultSourceHost()
	}

	s.sourceMu.Lock()
	s.sourceHost = host
	s.sourcePort = port
	s.sourceMu.Unlock()
}
