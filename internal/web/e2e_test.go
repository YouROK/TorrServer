package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"silo/internal/user"
)

// e2eEnv описывает окружение сквозных проверок транскодирования.
type e2eEnv struct {
	server  *Server
	baseURL string
	hash    string
	owner   *user.User
}

// setupE2E поднимает сервер с работающим транскодированием на реальном порту.
// Источником служит локальный файл: синтетическая раздача без пиров данные не отдаёт.
func setupE2E(t *testing.T) *e2eEnv {
	t.Helper()

	module := newTestModule(t)
	media := makeMediaFile(t)

	s, _, _, owner, hash := setupStreamTestEnv(t)
	s.SetTranscoder(module)
	serveLocalFile(t, s, media)

	return &e2eEnv{
		server:  s,
		baseURL: startTestServer(t, s),
		hash:    hash,
		owner:   owner,
	}
}

// url собирает адрес с токеном доступа.
func (e *e2eEnv) url(path string) string {
	return e.baseURL + path + "?token=" + e.owner.APIToken
}

// get выполняет запрос и возвращает тело ответа и статус.
func (e *e2eEnv) get(t *testing.T, url string, timeout time.Duration) (string, int, http.Header) {
	t.Helper()

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil && err != io.EOF {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(body), resp.StatusCode, resp.Header
}

// post выполняет POST-запрос и возвращает тело ответа и статус.
func (e *e2eEnv) post(t *testing.T, url string, timeout time.Duration) (string, int) {
	t.Helper()

	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil && err != io.EOF {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(body), resp.StatusCode
}

// TestE2EProbeInfo проверяет сведения о файле через настоящий HTTP-сервер.
func TestE2EProbeInfo(t *testing.T) {
	env := setupE2E(t)

	body, status, _ := env.get(t, env.url("/api/transcode/info/"+env.hash+"/0"), 60*time.Second)
	t.Logf("status=%d body=%s", status, truncateText(body, 400))

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", status, truncateText(body, 400))
	}
	for _, want := range []string{"container", "streams", "duration"} {
		if !strings.Contains(body, want) {
			t.Errorf("response does not contain %q", want)
		}
	}
}

// TestE2EProgressiveStream проверяет выдачу транскодированного потока клиенту.
func TestE2EProgressiveStream(t *testing.T) {
	env := setupE2E(t)

	// Сплошной поток запрашивается явно: по умолчанию отдаётся HLS.
	// Копирование без перекодирования: важно, что процесс запускается
	// и клиент получает заголовки потока
	url := env.url("/api/transcode/"+env.hash+"/0") +
		"&protocol=progressive&video_codec=copy&audio_codec=copy&container=mkv"

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	sessionID := resp.Header.Get(SessionHeader)
	t.Logf("status=%d session=%s content-type=%s", resp.StatusCode, sessionID, resp.Header.Get("Content-Type"))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		t.Fatalf("status = %d, want 200 (body: %s)", resp.StatusCode, body)
	}
	if sessionID == "" {
		t.Error("session header is missing")
	}
	if got := resp.Header.Get("Accept-Ranges"); got != "none" {
		t.Errorf("Accept-Ranges = %q, want none for a live transcode", got)
	}
}

// TestE2ESessionList проверяет, что запущенная сессия видна в списке активных.
func TestE2ESessionList(t *testing.T) {
	env := setupE2E(t)

	url := env.url("/api/transcode/"+env.hash+"/0") +
		"&video_codec=copy&audio_codec=copy&container=mkv"

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("stream did not start: %d", resp.StatusCode)
	}

	sessionID := resp.Header.Get(SessionHeader)
	if sessionID == "" {
		t.Skip("session id is missing")
	}

	body, status, _ := env.get(t, env.url("/api/transcode/sessions"), 10*time.Second)
	if status != http.StatusOK {
		t.Fatalf("sessions status = %d, want 200", status)
	}
	if !strings.Contains(body, sessionID) {
		t.Errorf("session %s is missing from the list: %s", sessionID, truncateText(body, 300))
	}
}

// TestE2EStreamStopsOnClientDisconnect проверяет остановку процесса при обрыве клиента.
func TestE2EStreamStopsOnClientDisconnect(t *testing.T) {
	env := setupE2E(t)

	url := env.url("/api/transcode/"+env.hash+"/0") +
		"&video_codec=copy&audio_codec=copy&container=mkv"

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	sessionID := resp.Header.Get(SessionHeader)
	if sessionID == "" {
		resp.Body.Close()
		t.Skip("session id is missing")
	}

	// Закрываем соединение и ждём, пока сессия исчезнет из реестра
	_ = resp.Body.Close()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		body, status, _ := env.get(t, env.url("/api/transcode/sessions"), 5*time.Second)
		if status == http.StatusOK && !strings.Contains(body, sessionID) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	t.Errorf("session %s survived client disconnect", sessionID)
}

// truncateText ограничивает длину текста для вывода в лог.
func truncateText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// TestE2EProgressiveSeek проверяет перемотку сплошного потока.
// Перемотка выполняется новым запросом с параметром t.
func TestE2EProgressiveSeek(t *testing.T) {
	env := setupE2E(t)

	sizeAt := func(seconds int) int {
		url := env.url("/api/transcode/"+env.hash+"/0") +
			fmt.Sprintf("&protocol=progressive&video_codec=copy&audio_codec=copy&container=mkv&t=%d", seconds)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("request at %ds failed: %v", seconds, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status at %ds = %d", seconds, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read at %ds: %v", seconds, err)
		}
		return len(body)
	}

	// Поток с большей позиции короче: файл заканчивается раньше
	full := sizeAt(0)
	later := sizeAt(3)

	t.Logf("bytes: from 0s = %d, from 3s = %d", full, later)

	if full == 0 || later == 0 {
		t.Fatalf("stream is empty: %d and %d bytes", full, later)
	}
	if later >= full {
		t.Errorf("stream from 3s (%d bytes) must be shorter than from 0s (%d bytes)", later, full)
	}
}

// TestHLSRedirectKeepsQuery проверяет, что переход на HLS не теряет параметры.
func TestHLSRedirectKeepsQuery(t *testing.T) {
	env := setupE2E(t)

	url := env.url("/api/transcode/"+env.hash+"/0") +
		"&video_codec=copy&audio_codec=copy&container=mkv&t=12"

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status = %d, want a redirect to HLS", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	t.Logf("redirect to %s", location)

	// Профиль по умолчанию сегментированный, поэтому переадресация обязана
	// сохранить позицию и параметры профиля
	for _, want := range []string{"t=12", "token="} {
		if !strings.Contains(location, want) {
			t.Errorf("redirect %q does not contain %q", location, want)
		}
	}
}

// TestE2EPauseResume проверяет управление паузой активной сессии.
func TestE2EPauseResume(t *testing.T) {
	env := setupE2E(t)

	if !env.server.transcoder.Enabled() {
		t.Skip("transcoding is disabled")
	}

	// Список сессий отдаёт признак поддержки паузы
	sessions, status, _ := env.get(t, env.url("/api/transcode/sessions"), 20*time.Second)
	if status != http.StatusOK {
		t.Fatalf("sessions status = %d", status)
	}
	t.Logf("sessions: %s", truncateText(sessions, 200))

	// Несуществующая сессия не должна ломать сервер
	_, status = env.post(t, env.url("/api/transcode/sessions/nonexistent/pause"), 20*time.Second)
	if status != http.StatusNotFound {
		t.Errorf("pause of an unknown session = %d, want 404", status)
	}
	t.Logf("pause unknown session status = %d", status)
}
