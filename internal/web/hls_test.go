package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"silo/internal/config"
	"silo/internal/ffmpeg"
	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/profile"
)

// hlsEnv описывает окружение проверок HLS.
type hlsEnv struct {
	env *e2eEnv
}

// setupHLS поднимает сервер с HLS-профилем и локальным источником.
func setupHLS(t *testing.T) *hlsEnv {
	t.Helper()

	module := newTestModule(t)
	media := makeMediaFile(t)

	s, _, _, owner, hash := setupStreamTestEnv(t)
	s.SetTranscoder(module)
	serveLocalFile(t, s, media)

	base := &e2eEnv{
		server:  s,
		baseURL: startTestServer(t, s),
		hash:    hash,
		owner:   owner,
	}
	return &hlsEnv{env: base}
}

func TestHLSMasterPlaylist(t *testing.T) {
	h := setupHLS(t)

	url := h.env.url("/api/transcode/hls/" + h.env.hash + "/0/master.m3u8")
	url += "&protocol=hls&container=ts&video_codec=h264&audio_codec=aac&segment_length=2"

	body, status, header := h.env.get(t, url, 90*time.Second)
	t.Logf("status=%d content-type=%s body=%s", status, header.Get("Content-Type"), truncateText(body, 300))

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", status, body)
	}
	if ct := header.Get("Content-Type"); !strings.Contains(ct, "mpegurl") {
		t.Errorf("content type = %q, want a playlist type", ct)
	}
	if !strings.HasPrefix(body, "#EXTM3U") {
		t.Errorf("master playlist must start with EXTM3U:\n%s", body)
	}
	if !strings.Contains(body, "#EXT-X-STREAM-INF") {
		t.Errorf("master playlist has no variants:\n%s", body)
	}
	if !strings.Contains(body, "main.m3u8") {
		t.Errorf("master playlist does not point to the media playlist:\n%s", body)
	}
}

func TestHLSMediaPlaylist(t *testing.T) {
	h := setupHLS(t)

	masterURL := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&container=ts&video_codec=h264&audio_codec=aac&segment_length=2"

	body, status, _ := h.env.get(t, masterURL, 90*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d, want 200 (body: %s)", status, truncateText(body, 200))
	}

	variant := variantURI(t, body)
	playlistURL := h.env.baseURL + variant

	// Ждём, пока появится первый сегмент: плейлист создаётся вместе с ним
	var playlistBody string
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		playlistBody, status, _ = h.env.get(t, playlistURL, 30*time.Second)
		if status == http.StatusOK && strings.Contains(playlistBody, "#EXTINF") {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	t.Logf("status=%d playlist=%s", status, truncateText(playlistBody, 300))

	if status != http.StatusOK {
		t.Fatalf("playlist status = %d, want 200", status)
	}
	if !strings.Contains(playlistBody, "#EXT-X-TARGETDURATION") {
		t.Errorf("playlist has no target duration:\n%s", playlistBody)
	}
	if !strings.Contains(playlistBody, "/segment/") {
		t.Errorf("playlist does not point to our segment endpoint:\n%s", playlistBody)
	}
}

func TestHLSSegmentDelivery(t *testing.T) {
	h := setupHLS(t)

	segmentURL := h.startAndFirstSegment(t)

	resp, err := http.Get(segmentURL)
	if err != nil {
		t.Fatalf("segment request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		t.Fatalf("segment status = %d, want 200 (body: %s)", resp.StatusCode, body)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		t.Fatalf("failed to read segment: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("segment is empty")
	}
	// Сегмент mpegts начинается с синхробайта 0x47
	if data[0] != 0x47 {
		t.Errorf("segment does not look like mpegts: first byte 0x%02x", data[0])
	}
}

func TestHLSSegmentRejectsTraversal(t *testing.T) {
	h := setupHLS(t)

	// Создаём сессию через мастер-плейлист
	masterURL := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&container=ts&video_codec=h264&audio_codec=aac&segment_length=2"

	body, status, _ := h.env.get(t, masterURL, 90*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d", status)
	}

	variant := variantURI(t, body)
	sessionID := sessionIDFromVariant(t, variant)

	for _, name := range []string{"..%2f..%2fetc%2fpasswd", "....//secret.ts"} {
		url := h.env.baseURL + "/api/transcode/hls/session/" + sessionID + "/segment/" + name
		if !strings.Contains(url, "token=") {
			url += "?token=" + h.env.owner.APIToken
		}

		_, status, _ := h.env.get(t, url, 20*time.Second)
		if status == http.StatusOK {
			t.Errorf("traversal name %q was accepted", name)
		}
	}
}

func TestHLSUnknownSession(t *testing.T) {
	h := setupHLS(t)

	url := h.env.url("/api/transcode/hls/session/deadbeef/main.m3u8")
	_, status, _ := h.env.get(t, url, 15*time.Second)

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for an unknown session", status)
	}
}

// startAndFirstSegment запускает HLS и возвращает адрес первого сегмента.
func (h *hlsEnv) startAndFirstSegment(t *testing.T) string {
	t.Helper()

	masterURL := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&container=ts&video_codec=h264&audio_codec=aac&segment_length=2"

	body, status, _ := h.env.get(t, masterURL, 90*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d, want 200 (body: %s)", status, truncateText(body, 200))
	}

	variant := variantURI(t, body)
	playlistURL := h.env.baseURL + variant

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		playlist, status, _ := h.env.get(t, playlistURL, 30*time.Second)
		if status == http.StatusOK {
			if segment := firstSegmentURI(playlist); segment != "" {
				return h.env.baseURL + segment
			}
		}
		time.Sleep(300 * time.Millisecond)
	}

	t.Fatal("no segment appeared in the playlist")
	return ""
}

// variantURI достаёт адрес медиаплейлиста из главного плейлиста.
func variantURI(t *testing.T, master string) string {
	t.Helper()

	for _, line := range strings.Split(master, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "main.m3u8") {
			return line
		}
	}
	t.Fatalf("no variant found in the master playlist:\n%s", master)
	return ""
}

// firstSegmentURI достаёт адрес первого сегмента из медиаплейлиста.
func firstSegmentURI(playlist string) string {
	for _, line := range strings.Split(playlist, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "/segment/") {
			return line
		}
	}
	return ""
}

// sessionIDFromVariant вытаскивает идентификатор сессии из адреса плейлиста.
func sessionIDFromVariant(t *testing.T, variant string) string {
	t.Helper()

	_, rest, ok := strings.Cut(variant, "/hls/session/")
	if !ok {
		t.Fatalf("variant %q has no session id", variant)
	}
	id, _, _ := strings.Cut(rest, "/")
	return id
}

func TestRemoveSegmentsBefore(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"seg0.ts", "seg1.ts", "seg2.ts", "seg3.ts", "index.m3u8"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	removeSegmentsBefore(dir, 2)

	for _, name := range []string{"seg0.ts", "seg1.ts"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("%s must be removed", name)
		}
	}
	for _, name := range []string{"seg2.ts", "seg3.ts", "index.m3u8"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s must stay: %v", name, err)
		}
	}
}

func TestRemoveAllSegments(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"seg0.ts", "seg1.m4s", "index.m3u8", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	removeSegments(dir)

	for _, name := range []string{"seg0.ts", "seg1.m4s"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("%s must be removed", name)
		}
	}
	for _, name := range []string{"index.m3u8", "notes.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s must stay: %v", name, err)
		}
	}
}

func TestIsSegmentName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: "seg0.ts", want: true},
		{name: "seg1.m4s", want: true},
		{name: "index.m3u8", want: false},
		{name: "notes.txt", want: false},
		{name: ".hidden.ts", want: false},
	}

	for _, c := range cases {
		if got := isSegmentName(c.name); got != c.want {
			t.Errorf("isSegmentName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestSegmentTailFromConfig(t *testing.T) {
	s := &Server{cfg: config.DefaultConfig()}

	// Значение по умолчанию, когда настройка не задана
	s.cfg.FFmpeg.CacheSegments = 0
	if got := s.segmentTail(); got != ffmpeg.DefaultCacheSegments {
		t.Errorf("segment tail = %d, want %d", got, ffmpeg.DefaultCacheSegments)
	}

	// Значение из конфигурации
	s.cfg.FFmpeg.CacheSegments = 12
	if got := s.segmentTail(); got != 12 {
		t.Errorf("segment tail = %d, want 12", got)
	}
}

func TestSegmentTailWithoutConfig(t *testing.T) {
	s := &Server{}

	if got := s.segmentTail(); got != ffmpeg.DefaultCacheSegments {
		t.Errorf("segment tail = %d, want %d", got, ffmpeg.DefaultCacheSegments)
	}
}

func TestHLSFmp4Playlist(t *testing.T) {
	h := setupHLS(t)

	url := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&container=ts&video_codec=h264&audio_codec=aac&segment_type=fmp4&segment_length=2"

	body, status, _ := h.env.get(t, url, 90*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d, want 200 (body: %s)", status, truncateText(body, 200))
	}

	variant := variantURI(t, body)
	sessionID := sessionIDFromVariant(t, variant)
	playlistURL := h.env.baseURL + variant

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		playlist, st, _ := h.env.get(t, playlistURL, 30*time.Second)
		if st != http.StatusOK || !strings.Contains(playlist, "#EXTINF") {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		t.Logf("playlist:\n%s", truncateText(playlist, 500))

		if !strings.Contains(playlist, "#EXT-X-MAP") {
			t.Errorf("fmp4 playlist has no EXT-X-MAP:\n%s", playlist)
		}
		if !strings.Contains(playlist, "init.mp4") {
			t.Errorf("init segment is not referenced:\n%s", playlist)
		}
		// Сегменты fmp4 должны иметь расширение m4s, иначе плеер не поймёт формат
		if strings.Contains(playlist, ".ts") {
			t.Errorf("fmp4 playlist references mpegts segments:\n%s", playlist)
		}

		initURL := h.env.baseURL + "/api/transcode/hls/session/" + sessionID + "/segment/init.mp4?token=" + h.env.owner.APIToken
		_, initStatus, _ := h.env.get(t, initURL, 30*time.Second)
		if initStatus != http.StatusOK {
			t.Errorf("init segment status = %d, want 200", initStatus)
		}

		segment := firstSegmentURI(playlist)
		if segment == "" {
			t.Fatal("no segment in the playlist")
		}
		_, segStatus, _ := h.env.get(t, h.env.baseURL+segment, 30*time.Second)
		if segStatus != http.StatusOK {
			t.Errorf("segment status = %d, want 200", segStatus)
		}
		return
	}

	t.Fatal("no fmp4 playlist appeared")
}

func TestHLSMpegtsHasNoInitSegment(t *testing.T) {
	h := setupHLS(t)

	url := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&container=ts&video_codec=h264&audio_codec=aac&segment_length=2"

	body, status, _ := h.env.get(t, url, 90*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d", status)
	}

	variant := variantURI(t, body)
	sessionID := sessionIDFromVariant(t, variant)

	// У mpegts начального сегмента нет, запрос должен быстро получить отказ
	initURL := h.env.baseURL + "/api/transcode/hls/session/" + sessionID + "/segment/init.mp4?token=" + h.env.owner.APIToken
	_, initStatus, _ := h.env.get(t, initURL, 15*time.Second)

	if initStatus != http.StatusNotFound {
		t.Errorf("init segment status = %d, want 404 for a mpegts session", initStatus)
	}
}

func TestCodecsForUsesProfile(t *testing.T) {
	// Источник в HEVC, но транскодируем в H.264: плееру объявляются целевые кодеки
	prof := profile.Default()
	prof.Video.Codec = "h264"
	prof.Audio.Codec = "aac"

	got := codecsFor(prof)
	if !strings.Contains(got, "avc1") {
		t.Errorf("codecs = %q, want an avc1 entry", got)
	}
	if !strings.Contains(got, "mp4a") {
		t.Errorf("codecs = %q, want an mp4a entry", got)
	}
	if strings.Contains(got, "hvc1") {
		t.Errorf("codecs = %q, must not announce the source codec", got)
	}
	if strings.Contains(got, "ac-3") {
		t.Errorf("codecs = %q, must not announce the source audio", got)
	}
}

func TestCodecsForSkipsDisabledStreams(t *testing.T) {
	prof := profile.Default()
	prof.Video.Codec = "none"
	prof.Audio.Codec = "aac"

	got := codecsFor(prof)
	if strings.Contains(got, "none") {
		t.Errorf("codecs = %q, must not contain a placeholder", got)
	}
	if !strings.Contains(got, "mp4a") {
		t.Errorf("codecs = %q, want the audio codec", got)
	}
}

func TestSegmentContentType(t *testing.T) {
	ts := profile.Default()
	ts.HLS.SegmentType = "mpegts"

	if got := segmentContentType("seg0.ts", ts); got != "video/mp2t" {
		t.Errorf("mpegts segment type = %q, want video/mp2t", got)
	}

	fmp4 := profile.Default()
	fmp4.HLS.SegmentType = "fmp4"
	fmp4.HLS.InitFilename = "init.mp4"

	if got := segmentContentType("seg0.m4s", fmp4); got != "video/iso.segment" {
		t.Errorf("fmp4 segment type = %q, want video/iso.segment", got)
	}
	if got := segmentContentType("init.mp4", fmp4); got != "video/mp4" {
		t.Errorf("init segment type = %q, want video/mp4", got)
	}
}

// TestHLSMasterAnnouncesTargetCodecs проверяет, что главный плейлист объявляет
// кодеки результата, а не источника. Источник здесь h264, а цель - hevc.
func TestHLSMasterAnnouncesTargetCodecs(t *testing.T) {
	h := setupHLS(t)

	url := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&video_codec=hevc&audio_codec=aac"

	body, status, _ := h.env.get(t, url, 60*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d, want 200 (body: %s)", status, truncateText(body, 200))
	}

	t.Logf("master:\n%s", body)

	if !strings.Contains(body, "hvc1") {
		t.Errorf("master does not announce the target video codec:\n%s", body)
	}
	if strings.Contains(body, "avc1") {
		t.Errorf("master announces the source codec instead of the target:\n%s", body)
	}
}

// TestHLSSegmentContentType проверяет тип сегмента mpegts.
func TestHLSSegmentContentType(t *testing.T) {
	h := setupHLS(t)

	url := h.env.url("/api/transcode/hls/"+h.env.hash+"/0/master.m3u8") +
		"&protocol=hls&video_codec=h264&audio_codec=aac&segment_length=2"

	body, status, _ := h.env.get(t, url, 60*time.Second)
	if status != http.StatusOK {
		t.Fatalf("master status = %d", status)
	}

	variant := variantURI(t, body)
	playlistURL := h.env.baseURL + variant

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		playlist, st, _ := h.env.get(t, playlistURL, 30*time.Second)
		if st != http.StatusOK || !strings.Contains(playlist, "#EXTINF") {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		segment := firstSegmentURI(playlist)
		if segment == "" {
			t.Fatal("no segment in the playlist")
		}

		// Расширение .ts закреплено и за форматом переводов,
		// поэтому тип обязан задаваться явно
		_, segStatus, header := h.env.get(t, h.env.baseURL+segment, 30*time.Second)
		if segStatus != http.StatusOK {
			t.Fatalf("segment status = %d", segStatus)
		}
		if got := header.Get("Content-Type"); !strings.HasPrefix(got, "video/mp2t") {
			t.Errorf("segment content type = %q, want video/mp2t", got)
		}
		return
	}

	t.Fatal("no playlist appeared")
}

func TestProfileFingerprintDistinguishesSettings(t *testing.T) {
	base := profile.Default()

	// Одинаковые настройки дают одинаковый отпечаток
	first := base
	second := base
	if profileFingerprint(first) != profileFingerprint(second) {
		t.Error("identical profiles must share a fingerprint")
	}

	// Разные разрешения дают разные отпечатки
	changed := base
	changed.Video.MaxHeight = 1080
	if profileFingerprint(base) == profileFingerprint(changed) {
		t.Error("different resolution must change the fingerprint")
	}

	// Режим тонемаппинга тоже учитывается
	tonemapped := base
	tonemapped.Video.TonemapMode = args.TonemapOn
	if profileFingerprint(base) == profileFingerprint(tonemapped) {
		t.Error("different tonemap mode must change the fingerprint")
	}

	// Тип сегментов влияет на плейлист
	fmp4 := base
	fmp4.HLS.SegmentType = "fmp4"
	if profileFingerprint(base) == profileFingerprint(fmp4) {
		t.Error("different segment type must change the fingerprint")
	}
}

func TestProfileFingerprintIgnoresMetadata(t *testing.T) {
	base := profile.Default()

	// Имя, владелец и отметки времени не влияют на результат кодирования
	renamed := base
	renamed.Name = "Другое имя"
	renamed.OwnerID = "other-user"
	renamed.IsDefault = true
	renamed.CreatedAt = base.CreatedAt.Add(time.Hour)
	renamed.UpdatedAt = base.UpdatedAt.Add(2 * time.Hour)

	if profileFingerprint(base) != profileFingerprint(renamed) {
		t.Error("metadata must not affect the fingerprint")
	}
}

func TestHLSKeyIncludesFingerprint(t *testing.T) {
	keyA := hlsKey("hash", 0, "user", "abc123")
	keyB := hlsKey("hash", 0, "user", "def456")

	if keyA == keyB {
		t.Error("different fingerprints must produce different keys")
	}
	// Разные файлы и пользователи тоже различаются
	if keyA == hlsKey("hash", 1, "user", "abc123") {
		t.Error("different file index must produce a different key")
	}
	if keyA == hlsKey("hash", 0, "other", "abc123") {
		t.Error("different user must produce a different key")
	}
}
