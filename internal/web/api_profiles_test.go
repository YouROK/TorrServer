package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"silo/internal/config"
	"silo/internal/database"
	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/profile"
	"silo/internal/ffmpeg/source"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

// profileEnv описывает окружение проверок профилей через HTTP.
type profileEnv struct {
	server *Server
	router *gin.Engine
	owner  *user.User
	admin  *user.User
	user   *user.User
}

// setupProfileEnv поднимает сервер с сервисом профилей и тремя ролями.
func setupProfileEnv(t *testing.T) *profileEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open(t.TempDir() + "/profiles_web.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := config.DefaultConfig()
	// Пароль владельца включает проверку токенов, иначе любой запрос
	// проходит как владелец и различить роли невозможно
	cfg.Auth.OwnerPassword = "secret"
	userSvc := user.NewService(user.NewStore(db), cfg)

	owner, _, err := userSvc.Login("owner", cfg.Auth.OwnerPassword)
	if err != nil {
		t.Fatalf("failed to authenticate owner: %v", err)
	}

	admin, err := userSvc.CreateUser(owner, "admin1", "pass1", user.RankAdmin, user.Limits{})
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	regular, err := userSvc.CreateUser(owner, "user1", "pass2", user.RankUser, user.Limits{})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	srv := &Server{
		cfg:            cfg,
		userSvc:        userSvc,
		router:         gin.New(),
		streamTracker:  NewStreamTracker(),
		sourceRegistry: source.NewRegistry(0),
		profiles:       profile.NewService(profile.NewStore(db)),
	}
	srv.registerRoutes()

	return &profileEnv{server: srv, router: srv.router, owner: owner, admin: admin, user: regular}
}

// call выполняет запрос к маршрутизатору от имени пользователя.
func (e *profileEnv) call(t *testing.T, method, path string, actor *user.User, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if actor != nil {
		req.AddCookie(&http.Cookie{Name: CookieTokenName, Value: actor.APIToken})
	}

	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

// decodeProfile разбирает профиль из ответа.
func decodeProfile(t *testing.T, rec *httptest.ResponseRecorder) profile.Profile {
	t.Helper()

	var p profile.Profile
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("failed to decode profile: %v (body: %s)", err, rec.Body.String())
	}
	return p
}

func TestProfileCreateAndList(t *testing.T) {
	env := setupProfileEnv(t)

	newProfile := profile.Default()
	newProfile.ID = ""
	newProfile.Name = "Phone 720p"
	newProfile.IsDefault = false
	newProfile.Visibility = profile.VisibilityPrivate
	newProfile.Video.MaxHeight = 720

	rec := env.call(t, http.MethodPost, "/api/transcode/profiles", env.admin, newProfile)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", rec.Code, rec.Body.String())
	}

	created := decodeProfile(t, rec)
	if created.ID == "" {
		t.Fatal("created profile has no id")
	}
	if created.OwnerID != env.admin.ID {
		t.Errorf("owner = %q, want %q", created.OwnerID, env.admin.ID)
	}

	rec = env.call(t, http.MethodGet, "/api/transcode/profiles", env.admin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", rec.Code)
	}

	var list struct {
		Profiles []profile.Profile `json:"profiles"`
		All      []profile.Profile `json:"all"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if len(list.Profiles) != 1 {
		t.Fatalf("profiles = %d, want only the created one", len(list.Profiles))
	}
	if list.Profiles[0].ID != created.ID {
		t.Errorf("profile id = %q, want %q", list.Profiles[0].ID, created.ID)
	}
}

func TestProfileListHidesBuiltin(t *testing.T) {
	env := setupProfileEnv(t)

	rec := env.call(t, http.MethodGet, "/api/transcode/profiles", env.admin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", rec.Code)
	}

	var list struct {
		Profiles []profile.Profile `json:"profiles"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}

	// Встроенный профиль живёт вне базы и в списке не показывается:
	// владелец сам назначает, какой профиль использовать по умолчанию
	for _, p := range list.Profiles {
		if p.ID == profile.DefaultID {
			t.Error("builtin profile must not appear in the list")
		}
	}
}

func TestBuiltinProfileIsNotDefault(t *testing.T) {
	// Ни один профиль не выбран по умолчанию, пока владелец этого не сделает
	builtin := profile.Default()
	if builtin.IsDefault {
		t.Error("builtin profile must not be marked as default")
	}
	if builtin.Protocol != args.ProtocolHLS {
		t.Errorf("builtin protocol = %q, want hls", builtin.Protocol)
	}
}

func TestProfileCreateRequiresAdmin(t *testing.T) {
	env := setupProfileEnv(t)

	newProfile := profile.Default()
	newProfile.ID = ""
	newProfile.Name = "User profile"
	newProfile.IsDefault = false

	rec := env.call(t, http.MethodPost, "/api/transcode/profiles", env.user, newProfile)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403 for a regular user", rec.Code)
	}
}

func TestProfileCreateRejectsBadJSON(t *testing.T) {
	env := setupProfileEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/transcode/profiles", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: CookieTokenName, Value: env.admin.APIToken})

	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestProfileCreateRejectsInvalid(t *testing.T) {
	env := setupProfileEnv(t)

	bad := profile.Default()
	bad.ID = ""
	bad.Name = "Broken"
	bad.IsDefault = false
	bad.Video.Codec = ""

	rec := env.call(t, http.MethodPost, "/api/transcode/profiles", env.admin, bad)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestProfileUpdate(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Before", profile.VisibilityPrivate)

	patch := created
	patch.Name = "After"
	patch.Video.CRF = 18

	rec := env.call(t, http.MethodPut, "/api/transcode/profiles/"+created.ID, env.admin, patch)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	updated := decodeProfile(t, rec)
	if updated.Name != "After" {
		t.Errorf("name = %q, want After", updated.Name)
	}
	if updated.Video.CRF != 18 {
		t.Errorf("crf = %d, want 18", updated.Video.CRF)
	}
}

func TestProfileUpdateForeignDenied(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Admin only", profile.VisibilityPrivate)

	// Второй админ не может править чужой профиль
	other, err := env.server.userSvc.CreateUser(env.owner, "admin2", "pass3", user.RankAdmin, user.Limits{})
	if err != nil {
		t.Fatalf("failed to create second admin: %v", err)
	}

	patch := created
	patch.Name = "Hijacked"

	rec := env.call(t, http.MethodPut, "/api/transcode/profiles/"+created.ID, other, patch)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestProfileGetUnknown(t *testing.T) {
	env := setupProfileEnv(t)

	rec := env.call(t, http.MethodGet, "/api/transcode/profiles/missing", env.admin, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestProfileDelete(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Temp", profile.VisibilityPrivate)

	rec := env.call(t, http.MethodDelete, "/api/transcode/profiles/"+created.ID, env.admin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	rec = env.call(t, http.MethodGet, "/api/transcode/profiles/"+created.ID, env.admin, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 after deletion", rec.Code)
	}
}

func TestProfileDeleteBuiltin(t *testing.T) {
	env := setupProfileEnv(t)

	rec := env.call(t, http.MethodDelete, "/api/transcode/profiles/"+profile.DefaultID, env.owner, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for the builtin profile", rec.Code)
	}
}

func TestProfileSetDefault(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Global", profile.VisibilityPublic)

	// Админ не может назначить профиль по умолчанию
	rec := env.call(t, http.MethodPost, "/api/transcode/profiles/"+created.ID+"/default", env.admin, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403 for an admin", rec.Code)
	}

	// Владелец может
	rec = env.call(t, http.MethodPost, "/api/transcode/profiles/"+created.ID+"/default", env.owner, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	updated := decodeProfile(t, rec)
	if !updated.IsDefault {
		t.Error("profile was not marked as default")
	}
}

func TestProfileNameConflict(t *testing.T) {
	env := setupProfileEnv(t)

	createProfile(t, env, env.admin, "Same", profile.VisibilityPrivate)

	dup := profile.Default()
	dup.ID = ""
	dup.Name = "Same"
	dup.IsDefault = false

	rec := env.call(t, http.MethodPost, "/api/transcode/profiles", env.admin, dup)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestProfileRequiresAuth(t *testing.T) {
	env := setupProfileEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/transcode/profiles", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// createProfile создаёт профиль через API и возвращает его.
func createProfile(t *testing.T, env *profileEnv, actor *user.User, name string, visibility profile.Visibility) profile.Profile {
	t.Helper()

	p := profile.Default()
	p.ID = ""
	p.Name = name
	p.IsDefault = false
	p.Visibility = visibility

	rec := env.call(t, http.MethodPost, "/api/transcode/profiles", actor, p)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create profile %q: status %d (body: %s)", name, rec.Code, rec.Body.String())
	}
	return decodeProfile(t, rec)
}

func TestPreferredProfileFlow(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Preferred", profile.VisibilityPublic)

	// Пользователь выбирает доступный ему профиль
	rec := env.call(t, http.MethodPost, "/api/transcode/preferred-profile", env.user,
		map[string]string{"profile_id": created.ID})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	updated, err := env.server.userSvc.GetUserByID(env.user.ID)
	if err != nil {
		t.Fatalf("failed to read user: %v", err)
	}
	if updated.TranscodeProfile != created.ID {
		t.Errorf("preferred profile = %q, want %q", updated.TranscodeProfile, created.ID)
	}
}

func TestPreferredProfileRejectsPrivate(t *testing.T) {
	env := setupProfileEnv(t)

	// Приватный профиль админа недоступен обычному пользователю
	private := createProfile(t, env, env.admin, "Secret", profile.VisibilityPrivate)

	rec := env.call(t, http.MethodPost, "/api/transcode/preferred-profile", env.user,
		map[string]string{"profile_id": private.ID})
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestPreferredProfileEmptyResetsToDefault(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Temporary", profile.VisibilityPublic)

	rec := env.call(t, http.MethodPost, "/api/transcode/preferred-profile", env.user,
		map[string]string{"profile_id": created.ID})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// Пустое значение возвращает профиль по умолчанию
	rec = env.call(t, http.MethodPost, "/api/transcode/preferred-profile", env.user,
		map[string]string{"profile_id": ""})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	updated, err := env.server.userSvc.GetUserByID(env.user.ID)
	if err != nil {
		t.Fatalf("failed to read user: %v", err)
	}
	if updated.TranscodeProfile != "" {
		t.Errorf("preferred profile = %q, want it cleared", updated.TranscodeProfile)
	}
}

func TestResolveProfileUsesPreference(t *testing.T) {
	env := setupProfileEnv(t)

	created := createProfile(t, env, env.admin, "Chosen", profile.VisibilityPublic)
	if err := env.server.userSvc.SetTranscodeProfile(env.user, created.ID); err != nil {
		t.Fatalf("failed to set preference: %v", err)
	}

	actor := profile.Actor{ID: env.user.ID, Rank: int(env.user.Rank)}

	// Без явного идентификатора берётся выбранный пользователем профиль
	resolved, err := env.server.baseProfile(actor, "")
	if err != nil {
		t.Fatalf("baseProfile: %v", err)
	}
	if resolved.ID != created.ID {
		t.Errorf("resolved profile = %q, want %q", resolved.ID, created.ID)
	}
}

func TestResolveProfileExplicitOverridesPreference(t *testing.T) {
	env := setupProfileEnv(t)

	first := createProfile(t, env, env.admin, "First", profile.VisibilityPublic)
	second := createProfile(t, env, env.admin, "Second", profile.VisibilityPublic)

	if err := env.server.userSvc.SetTranscodeProfile(env.user, first.ID); err != nil {
		t.Fatalf("failed to set preference: %v", err)
	}

	actor := profile.Actor{ID: env.user.ID, Rank: int(env.user.Rank)}

	resolved, err := env.server.baseProfile(actor, second.ID)
	if err != nil {
		t.Fatalf("baseProfile: %v", err)
	}
	if resolved.ID != second.ID {
		t.Errorf("resolved profile = %q, want the explicit %q", resolved.ID, second.ID)
	}
}
