package web

import (
	"context"
	"fmt"

	"silo/internal/log"
	"silo/internal/user"
)

// fillDuration в фоне определяет длительность файла, если плеер её не сообщил.
// Файл разбирается один раз на пользователя, неудача не повторяется до перезапуска.
func (s *Server) fillDuration(u *user.User, hashHex string, fileIdx int) {
	if s.tasks == nil {
		return
	}
	if !s.probeEnabled() {
		s.durationOff.Do(func() {
			log.Infof("[Web] Duration detection is unavailable, progress keeps durations sent by players only: %s", s.probeReason())
		})
		return
	}

	ut, err := s.userSvc.GetUserTorrent(u.ID, hashHex)
	if err != nil {
		return
	}
	if p, ok := ut.Progress[fileIdx]; !ok || p.Duration > 0 {
		return
	}

	key := fmt.Sprintf("%s:%s:%d", u.ID, hashHex, fileIdx)
	if _, busy := s.durationProbes.LoadOrStore(key, struct{}{}); busy {
		return
	}

	s.tasks.Go(func(ctx context.Context) {
		lease, info, err := s.probeSource(ctx, u, hashHex, fileIdx)
		if err != nil {
			log.Warnf("[Web] Failed to probe duration of file %d in %s, not retrying until restart: %v", fileIdx, hashHex, err)
			return
		}
		s.sourceRegistry.Release(lease.ID)

		if info.Duration <= 0 {
			log.Warnf("[Web] No duration found for file %d in %s, not retrying until restart", fileIdx, hashHex)
			return
		}
		if err := s.userSvc.SetFileDuration(u, hashHex, fileIdx, info.Duration); err != nil {
			log.Errorf("[Web] Failed to save duration of file %d in %s: %v", fileIdx, hashHex, err)
			s.durationProbes.Delete(key)
			return
		}
		log.Debugf("[Web] Detected duration %.0fs for file %d in %s", info.Duration, fileIdx, hashHex)
		s.durationProbes.Delete(key)
	})
}
