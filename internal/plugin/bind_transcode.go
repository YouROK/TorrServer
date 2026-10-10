package plugin

import (
	"context"
	"fmt"
	"time"

	"silo/internal/ffmpeg"
	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/media"
	"silo/internal/ffmpeg/profile"
	"silo/internal/user"

	"github.com/dop251/goja"
)

// createTranscodeModule собирает объект ts.transcode для плагинов.
func (rt *JSRuntime) createTranscodeModule(module *ffmpeg.FFmpeg, profiles *profile.Service, prober ProbeFunc) *goja.Object {
	obj := rt.vm.NewObject()

	// resolveActor превращает идентификатор пользователя в права для профилей
	resolveActor := func(call goja.FunctionCall) profile.Actor {
		userID := call.Argument(0).String()
		if userID == "" {
			userID = "owner"
		}

		u, err := rt.userSvc.GetUserByID(userID)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("user not found: %s", userID)))
		}
		return actorOf(u)
	}

	requireProfiles := func() *profile.Service {
		if profiles == nil {
			panic(rt.vm.ToValue("profiles service is not linked"))
		}
		return profiles
	}

	requireModule := func() *ffmpeg.FFmpeg {
		if module == nil {
			panic(rt.vm.ToValue("transcoding module is not linked"))
		}
		return module
	}

	// ts.transcode.info() - состояние модуля транскодирования
	obj.Set("info", func(call goja.FunctionCall) goja.Value {
		if module == nil {
			return rt.vm.ToValue(map[string]any{"enabled": false, "reason": "module is not linked"})
		}
		return rt.vm.ToValue(module.Info())
	})

	// ts.transcode.sessions() - активные сессии
	obj.Set("sessions", func(call goja.FunctionCall) goja.Value {
		if module == nil {
			return rt.vm.ToValue([]any{})
		}
		return rt.vm.ToValue(module.Jobs())
	})

	// ts.transcode.stop(sessionID) - остановка сессии
	obj.Set("stop", func(call goja.FunctionCall) goja.Value {
		id := call.Argument(0).String()
		if id == "" {
			panic(rt.vm.ToValue("session id is required"))
		}
		if !requireModule().Stop(id) {
			return rt.vm.ToValue(false)
		}
		return rt.vm.ToValue(true)
	})

	// ts.transcode.probe(userID, hash, fileIdx) - разбор файла раздачи
	obj.Set("probe", func(call goja.FunctionCall) goja.Value {
		if prober == nil {
			panic(rt.vm.ToValue("probe is not available"))
		}

		userID := call.Argument(0).String()
		if userID == "" {
			userID = "owner"
		}
		hash := call.Argument(1).String()
		fileIdx := int(call.Argument(2).ToInteger())

		if hash == "" {
			panic(rt.vm.ToValue("torrent hash is required"))
		}

		// Разбор обращается к раздаче по сети, поэтому ограничиваем время
		ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
		defer cancel()

		info, err := prober(ctx, userID, hash, fileIdx)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to probe file: %v", err)))
		}

		// Структура передаётся как объект с именами полей из JSON,
		// иначе плагин увидит имена Go-полей
		return rt.vm.ToValue(mediaInfoToMap(info))
	})

	// ts.transcode.profiles(userID) - доступные пользователю профили
	obj.Set("profiles", func(call goja.FunctionCall) goja.Value {
		actor := resolveActor(call)
		list, err := requireProfiles().List(actor)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to list profiles: %v", err)))
		}
		return rt.vm.ToValue(list)
	})

	// ts.transcode.profile(userID, id) - один профиль
	obj.Set("profile", func(call goja.FunctionCall) goja.Value {
		actor := resolveActor(call)
		id := call.Argument(1).String()

		p, err := requireProfiles().Get(actor, id)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to get profile: %v", err)))
		}
		return rt.vm.ToValue(p)
	})

	// ts.transcode.createProfile(userID, options) - создание профиля
	obj.Set("createProfile", func(call goja.FunctionCall) goja.Value {
		actor := resolveActor(call)

		p, err := profileFromValue(call.Argument(1))
		if err != nil {
			panic(rt.vm.ToValue(err.Error()))
		}

		created, err := requireProfiles().Create(actor, p)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to create profile: %v", err)))
		}
		return rt.vm.ToValue(created)
	})

	// ts.transcode.updateProfile(userID, id, options) - изменение профиля
	obj.Set("updateProfile", func(call goja.FunctionCall) goja.Value {
		actor := resolveActor(call)
		id := call.Argument(1).String()

		p, err := profileFromValue(call.Argument(2))
		if err != nil {
			panic(rt.vm.ToValue(err.Error()))
		}

		updated, err := requireProfiles().Update(actor, id, p)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to update profile: %v", err)))
		}
		return rt.vm.ToValue(updated)
	})

	// ts.transcode.deleteProfile(userID, id) - удаление профиля
	obj.Set("deleteProfile", func(call goja.FunctionCall) goja.Value {
		actor := resolveActor(call)
		id := call.Argument(1).String()

		if err := requireProfiles().Delete(actor, id); err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("failed to delete profile: %v", err)))
		}
		return goja.Undefined()
	})

	// ts.transcode.defaultProfile() - профиль по умолчанию
	obj.Set("defaultProfile", func(call goja.FunctionCall) goja.Value {
		if profiles == nil {
			p := profile.Default()
			return rt.vm.ToValue(&p)
		}
		return rt.vm.ToValue(profiles.Default())
	})

	return obj
}

// mediaInfoToMap переводит сведения о файле в объект для плагина.
// Имена полей повторяют JSON-представление, чтобы плагин видел привычные ключи.
func mediaInfoToMap(info *media.MediaInfo) map[string]any {
	if info == nil {
		return nil
	}

	streams := make([]map[string]any, 0, len(info.Streams))
	for _, s := range info.Streams {
		streams = append(streams, map[string]any{
			"index":          s.Index,
			"kind":           string(s.Kind),
			"codec":          s.Codec,
			"profile":        s.Profile,
			"width":          s.Width,
			"height":         s.Height,
			"pix_fmt":        s.PixFmt,
			"bit_depth":      s.BitDepth,
			"framerate":      s.Framerate,
			"bitrate":        s.Bitrate,
			"channels":       s.Channels,
			"channel_layout": s.ChannelLayout,
			"sample_rate":    s.SampleRate,
			"language":       s.Language,
			"title":          s.Title,
			"default":        s.Default,
			"forced":         s.Forced,
			"is_text_sub":    s.IsTextSub,
		})
	}

	return map[string]any{
		"format_name": info.FormatName,
		"container":   info.Container,
		"duration":    info.Duration,
		"size":        info.Size,
		"bitrate":     info.Bitrate,
		"streams":     streams,
		"has_video":   info.HasVideo(),
		"has_audio":   info.HasAudio(),
	}
}

// probeTimeout ограничивает время разбора файла раздачи.
const probeTimeout = 60 * time.Second

// ProbeFunc описывает разбор файла раздачи, который предоставляет веб-слой.
// Функция передаётся сюда, чтобы модуль плагинов не зависел от веб-пакета.
type ProbeFunc func(ctx context.Context, userID, hash string, fileIdx int) (*media.MediaInfo, error)

// actorOf собирает права пользователя для операций с профилями.
func actorOf(u *user.User) profile.Actor {
	if u == nil {
		return profile.Actor{}
	}
	return profile.Actor{
		ID:    u.ID,
		Rank:  int(u.Rank),
		Admin: u.Rank >= user.RankAdmin,
		Owner: u.Rank >= user.RankOwner,
	}
}

// profileFromValue собирает профиль из объекта плагина.
func profileFromValue(value goja.Value) (*profile.Profile, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, fmt.Errorf("profile options are required")
	}

	raw, ok := value.Export().(map[string]any)
	if !ok {
		return nil, fmt.Errorf("profile options must be an object")
	}

	p := profile.Default()
	p.IsDefault = false
	p.Name = stringField(raw, "name")
	if p.Name == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	if v := stringField(raw, "protocol"); v != "" {
		p.Protocol = args.Protocol(v)
	}
	if v := stringField(raw, "container"); v != "" {
		p.Container = v
	}
	if v := stringField(raw, "visibility"); v != "" {
		p.Visibility = profile.Visibility(v)
	}
	if v, ok := intField(raw, "rank_required"); ok {
		p.RankRequired = v
	}
	if v, ok := boolField(raw, "is_default"); ok {
		p.IsDefault = v
	}
	if v, ok := boolField(raw, "allow_video_copy"); ok {
		p.AllowVideoCopy = v
	}
	if v, ok := boolField(raw, "allow_audio_copy"); ok {
		p.AllowAudioCopy = v
	}

	if video, ok := raw["video"].(map[string]any); ok {
		applyVideoOptions(&p.Video, video)
	}
	if audio, ok := raw["audio"].(map[string]any); ok {
		applyAudioOptions(&p.Audio, audio)
	}
	if hlsOptions, ok := raw["hls"].(map[string]any); ok {
		if v, ok := intField(hlsOptions, "segment_length"); ok {
			p.HLS.SegmentLength = v
		}
		if v := stringField(hlsOptions, "segment_type"); v != "" {
			p.HLS.SegmentType = v
		}
		if v, ok := intField(hlsOptions, "list_size"); ok {
			p.HLS.ListSize = v
		}
	}

	p.Normalize()
	return &p, nil
}

// applyVideoOptions переносит параметры видео из объекта плагина.
func applyVideoOptions(dst *args.VideoOptions, raw map[string]any) {
	if v := stringField(raw, "codec"); v != "" {
		dst.Codec = v
	}
	if v := stringField(raw, "encoder"); v != "" {
		dst.Encoder = v
	}
	if v := stringField(raw, "hwaccel"); v != "" {
		dst.HWAccel = v
	}
	if v := stringField(raw, "preset"); v != "" {
		dst.Preset = v
	}
	if v := stringField(raw, "profile"); v != "" {
		dst.Profile = v
	}
	if v := stringField(raw, "level"); v != "" {
		dst.Level = v
	}
	if v := stringField(raw, "pix_fmt"); v != "" {
		dst.PixFmt = v
	}
	if v, ok := intField(raw, "crf"); ok {
		dst.CRF = v
	}
	if v, ok := intField(raw, "bitrate_kbps"); ok {
		dst.BitrateKbps = v
	}
	if v, ok := intField(raw, "max_bitrate_kbps"); ok {
		dst.MaxBitrateKbps = v
	}
	if v, ok := intField(raw, "max_width"); ok {
		dst.MaxWidth = v
	}
	if v, ok := intField(raw, "max_height"); ok {
		dst.MaxHeight = v
	}
}

// applyAudioOptions переносит параметры аудио из объекта плагина.
func applyAudioOptions(dst *args.AudioOptions, raw map[string]any) {
	if v := stringField(raw, "codec"); v != "" {
		dst.Codec = v
	}
	if v := stringField(raw, "encoder"); v != "" {
		dst.Encoder = v
	}
	if v, ok := intField(raw, "bitrate_kbps"); ok {
		dst.BitrateKbps = v
	}
	if v, ok := intField(raw, "channels"); ok {
		dst.Channels = v
	}
	if v, ok := intField(raw, "sample_rate"); ok {
		dst.SampleRate = v
	}
}

// stringField читает строковое поле объекта.
func stringField(raw map[string]any, key string) string {
	if v, ok := raw[key].(string); ok {
		return v
	}
	return ""
}

// intField читает целочисленное поле объекта.
func intField(raw map[string]any, key string) (int, bool) {
	switch v := raw[key].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// boolField читает логическое поле объекта.
func boolField(raw map[string]any, key string) (bool, bool) {
	if v, ok := raw[key].(bool); ok {
		return v, true
	}
	return false, false
}
