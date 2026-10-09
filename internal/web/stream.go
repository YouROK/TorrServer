package web

import (
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"silo/internal/log"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// handleStream реализует HTTP-стриминг видео с поддержкой Seek и DLNA
func (s *Server) handleStream(c *gin.Context) {
	hashHex := c.Param("hash")
	fileIdxStr := c.Param("fileIdx")

	fileIdx, err := strconv.Atoi(fileIdxStr)
	if err != nil || fileIdx < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file index"})
		return
	}

	val, _ := c.Get("user")
	currentUser := val.(*user.User)

	// Пробуждаем торрент (если он спит в БД) и получаем поток байт
	reader, fileStat, err := s.torrentMgr.GetStreamReader(currentUser, hashHex, fileIdx)
	if err != nil {
		log.Errorf("[Web] Failed to get stream reader for %s/%d: %v", hashHex, fileIdx, err)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "torrent or file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open stream"})
		return
	}

	// Гарантируем закрытие ридера и отметку о просмотре
	defer func() {
		_ = reader.Close()
		if markErr := s.torrentMgr.SetFileViewed(currentUser, hashHex, fileIdx, true); markErr != nil {
			log.Warnf("[Web] Failed to mark file %d as viewed: %v", fileIdx, markErr)
		}
	}()

	// Определяем MIME-тип по расширению файла
	ext := strings.ToLower(path.Ext(fileStat.Path))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Устанавливаем заголовки для DLNA-плееров, Smart TV и браузеров
	c.Header("transferMode.dlna.org", "Streaming")
	c.Header("contentFeatures.dlna.org", "DLNA.ORG_OP=01;DLNA.ORG_CI=0")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", contentType)

	// Регистрируем поток в трекере: админка видит, кто и с какого IP смотрит.
	// Берём реальный адрес соединения (RemoteIP), а не ClientIP: у gin по умолчанию
	// все прокси считаются доверенными, поэтому ClientIP подставляется из заголовка
	// X-Forwarded-For от кого угодно. Оба адреса показываем отдельно.
	handle := s.streamTracker.Open(StreamOpen{
		UserID:      currentUser.ID,
		Hash:        hashHex,
		FileIdx:     fileIdx,
		FileName:    fileStat.Name,
		ClientIP:    c.RemoteIP(),
		ForwardedIP: forwardedIP(c),
		UserAgent:   c.Request.UserAgent(),
	})
	defer handle.Close()

	// Оборачиваем writer, чтобы считать реально отданные байты (для скорости).
	c.Writer = &countingWriter{ResponseWriter: c.Writer, handle: handle}

	// Делегируем обработку Range-запросов (Seek) стандартной библиотеке Go.
	// http.ServeContent сам распарсит заголовок Range, сделает Seek и отдаст нужный кусок.
	http.ServeContent(c.Writer, c.Request, fileStat.Path, time.Now(), newBufferedStreamReader(reader, 1<<20))
}

// forwardedIP возвращает адрес клиента из заголовков reverse-proxy.
// Нужен, когда Silo стоит за прокси: реальный клиент будет здесь,
// а не в адресе TCP-соединения.
func forwardedIP(c *gin.Context) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		raw := strings.TrimSpace(c.GetHeader(header))
		if raw == "" {
			continue
		}

		// X-Forwarded-For может содержать цепочку: client, proxy1, proxy2
		first := strings.TrimSpace(strings.Split(raw, ",")[0])
		if first != "" {
			return first
		}
	}
	return ""
}
