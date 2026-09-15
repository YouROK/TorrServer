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

	// 1. Пробуждаем торрент (если он спит в БД) и получаем поток байт
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

	// 2. Гарантируем закрытие ридера и отметку о просмотре
	defer func() {
		_ = reader.Close()
		if markErr := s.torrentMgr.SetFileViewed(currentUser, hashHex, fileIdx, true); markErr != nil {
			log.Warnf("[Web] Failed to mark file %d as viewed: %v", fileIdx, markErr)
		}
	}()

	// 3. Определяем MIME-тип по расширению файла
	ext := strings.ToLower(path.Ext(fileStat.Path))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 4. Устанавливаем заголовки для DLNA-плееров, Smart TV и браузеров
	c.Header("transferMode.dlna.org", "Streaming")
	c.Header("contentFeatures.dlna.org", "DLNA.ORG_OP=01;DLNA.ORG_CI=0")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", contentType)

	// 5. Делегируем обработку Range-запросов (Seek) стандартной библиотеке Go.
	// http.ServeContent сам распарсит заголовок Range, сделает Seek и отдаст нужный кусок.
	http.ServeContent(c.Writer, c.Request, fileStat.Path, time.Now(), reader)
}
