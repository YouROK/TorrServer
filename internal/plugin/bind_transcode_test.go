package plugin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/media"
	"silo/internal/ffmpeg/profile"
	"silo/internal/user"

	"github.com/dop251/goja"
)

// newRuntimeForTranscode создаёт рантайм для проверок привязки транскодинга.
func newRuntimeForTranscode(t *testing.T, svc *profile.Service) *JSRuntime {
	t.Helper()

	vm := goja.New()
	rt := &JSRuntime{
		pluginID: "tester",
		vm:       vm,
		userSvc:  nil,
	}

	obj := rt.createTranscodeModule(nil, svc, nil)
	vm.Set("transcode", obj)
	return rt
}

func TestTranscodeModuleWithoutDependencies(t *testing.T) {
	rt := newRuntimeForTranscode(t, nil)

	info, err := rt.vm.RunString(`JSON.stringify(transcode.info())`)
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.String() == "" || info.String() == "{}" {
		t.Errorf("info = %s, want a disabled report", info.String())
	}

	sessions, err := rt.vm.RunString(`transcode.sessions().length`)
	if err != nil {
		t.Fatalf("sessions: %v", err)
	}
	if sessions.ToInteger() != 0 {
		t.Errorf("sessions = %d, want 0", sessions.ToInteger())
	}
}

func TestTranscodeProfileFromValue(t *testing.T) {
	rt := newRuntimeForTranscode(t, nil)

	value, err := rt.vm.RunString(`({
		name: "Phone",
		protocol: "progressive",
		container: "mkv",
		visibility: "public",
		rank_required: 10,
		video: {codec: "h264", crf: 21, preset: "fast", max_height: 720},
		audio: {codec: "aac", bitrate_kbps: 160, channels: 2},
		hls: {segment_length: 6}
	})`)
	if err != nil {
		t.Fatalf("failed to build value: %v", err)
	}

	p, err := profileFromValue(value)
	if err != nil {
		t.Fatalf("profileFromValue: %v", err)
	}

	if p.Name != "Phone" {
		t.Errorf("name = %q, want Phone", p.Name)
	}
	if p.Container != "mkv" {
		t.Errorf("container = %q, want mkv", p.Container)
	}
	if p.Visibility != profile.VisibilityPublic {
		t.Errorf("visibility = %q, want public", p.Visibility)
	}
	if p.RankRequired != 10 {
		t.Errorf("rank = %d, want 10", p.RankRequired)
	}
	if p.Video.Codec != "h264" || p.Video.CRF != 21 {
		t.Errorf("video = %+v, want h264 with crf 21", p.Video)
	}
	if p.Video.MaxHeight != 720 {
		t.Errorf("max height = %d, want 720", p.Video.MaxHeight)
	}
	if p.Audio.BitrateKbps != 160 || p.Audio.Channels != 2 {
		t.Errorf("audio = %+v, want 160 kbps stereo", p.Audio)
	}
	if p.HLS.SegmentLength != 6 {
		t.Errorf("segment length = %d, want 6", p.HLS.SegmentLength)
	}
	if p.IsDefault {
		t.Error("plugin profiles must not be default by accident")
	}
}

func TestTranscodeProfileRequiresName(t *testing.T) {
	rt := newRuntimeForTranscode(t, nil)

	value, err := rt.vm.RunString(`({container: "mp4"})`)
	if err != nil {
		t.Fatalf("failed to build value: %v", err)
	}

	if _, err := profileFromValue(value); err == nil {
		t.Error("expected an error for a profile without a name")
	}
}

func TestTranscodeProfileRejectsMissingOptions(t *testing.T) {
	if _, err := profileFromValue(goja.Undefined()); err == nil {
		t.Error("expected an error for undefined options")
	}
	if _, err := profileFromValue(goja.Null()); err == nil {
		t.Error("expected an error for null options")
	}
}

func TestTranscodeProfileDefaultsAreFilled(t *testing.T) {
	rt := newRuntimeForTranscode(t, nil)

	value, err := rt.vm.RunString(`({name: "Minimal"})`)
	if err != nil {
		t.Fatalf("failed to build value: %v", err)
	}

	p, err := profileFromValue(value)
	if err != nil {
		t.Fatalf("profileFromValue: %v", err)
	}

	if p.Video.Codec == "" || p.Audio.Codec == "" {
		t.Errorf("codecs were not defaulted: %+v", p)
	}
	// Профиль без явного протокола получает общий по умолчанию
	if p.Protocol != args.ProtocolHLS {
		t.Errorf("protocol = %q, want hls", p.Protocol)
	}
	if p.HLS.SegmentLength <= 0 {
		t.Error("segment length was not defaulted")
	}
}

func TestActorOf(t *testing.T) {
	cases := []struct {
		name  string
		user  *user.User
		admin bool
		owner bool
	}{
		{name: "nil", user: nil},
		{name: "owner", user: &user.User{ID: "o", Rank: user.RankOwner}, admin: true, owner: true},
		{name: "admin", user: &user.User{ID: "a", Rank: user.RankAdmin}, admin: true},
		{name: "user", user: &user.User{ID: "u", Rank: user.RankUser}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actor := actorOf(c.user)
			if actor.Admin != c.admin {
				t.Errorf("admin = %v, want %v", actor.Admin, c.admin)
			}
			if actor.Owner != c.owner {
				t.Errorf("owner = %v, want %v", actor.Owner, c.owner)
			}
		})
	}
}

func TestTranscodeProbeWithoutFunc(t *testing.T) {
	rt := newRuntimeForTranscode(t, nil)

	// Без подключённой функции разбор невозможен
	_, err := rt.vm.RunString(`transcode.probe("owner", "aabb", 0)`)
	if err == nil {
		t.Fatal("expected an error when probe is not available")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("error = %v, want a message about the missing probe", err)
	}
}

func TestTranscodeProbeRequiresHash(t *testing.T) {
	rt := newRuntimeForProbe(t, func(ctx context.Context, userID, hash string, fileIdx int) (*media.MediaInfo, error) {
		return &media.MediaInfo{Container: "mkv"}, nil
	})

	_, err := rt.vm.RunString(`transcode.probe("owner", "", 0)`)
	if err == nil {
		t.Fatal("expected an error for an empty hash")
	}
	if !strings.Contains(err.Error(), "hash is required") {
		t.Errorf("error = %v, want a message about the hash", err)
	}
}

func TestTranscodeProbeReturnsInfo(t *testing.T) {
	var gotUser, gotHash string
	var gotIdx int

	rt := newRuntimeForProbe(t, func(ctx context.Context, userID, hash string, fileIdx int) (*media.MediaInfo, error) {
		gotUser, gotHash, gotIdx = userID, hash, fileIdx
		return &media.MediaInfo{
			Container: "mkv",
			Duration:  120.5,
			Streams: []media.Stream{
				{Index: 0, Kind: media.KindVideo, Codec: "h264", Width: 1920, Height: 1080},
			},
		}, nil
	})

	value, err := rt.vm.RunString(`transcode.probe("owner", "aabbcc", 2)`)
	if err != nil {
		t.Fatalf("probe: %v", err)
	}

	obj := value.ToObject(rt.vm)
	if got := obj.Get("container").String(); got != "mkv" {
		t.Errorf("container = %q, want mkv", got)
	}

	// Аргументы передаются в функцию как есть
	if gotUser != "owner" || gotHash != "aabbcc" || gotIdx != 2 {
		t.Errorf("probe called with (%q, %q, %d), want (owner, aabbcc, 2)", gotUser, gotHash, gotIdx)
	}
}

func TestTranscodeProbeDefaultUser(t *testing.T) {
	var gotUser string

	rt := newRuntimeForProbe(t, func(ctx context.Context, userID, hash string, fileIdx int) (*media.MediaInfo, error) {
		gotUser = userID
		return &media.MediaInfo{Container: "mp4"}, nil
	})

	if _, err := rt.vm.RunString(`transcode.probe("", "aabb", 0)`); err != nil {
		t.Fatalf("probe: %v", err)
	}
	if gotUser != "owner" {
		t.Errorf("user = %q, want owner as the default", gotUser)
	}
}

func TestTranscodeProbeErrorPropagates(t *testing.T) {
	rt := newRuntimeForProbe(t, func(ctx context.Context, userID, hash string, fileIdx int) (*media.MediaInfo, error) {
		return nil, errors.New("torrent not found")
	})

	_, err := rt.vm.RunString(`transcode.probe("owner", "aabb", 0)`)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "torrent not found") {
		t.Errorf("error = %v, want the underlying cause", err)
	}
}

// newRuntimeForProbe создаёт рантайм с подключённой функцией разбора.
func newRuntimeForProbe(t *testing.T, fn ProbeFunc) *JSRuntime {
	t.Helper()

	vm := goja.New()
	rt := &JSRuntime{pluginID: "tester", vm: vm}

	obj := rt.createTranscodeModule(nil, nil, fn)
	vm.Set("transcode", obj)
	return rt
}
