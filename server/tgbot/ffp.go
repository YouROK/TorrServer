package tgbot

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"server/ffprobe"
	"server/settings"
	"server/torr"

	"github.com/dustin/go-humanize"
	tele "gopkg.in/telebot.v4"
	ffp "gopkg.in/vansante/go-ffprobe.v2"
)

// TODO: Use internal API for ffp

func cmdFfp(c tele.Context) error {
	uid := c.Sender().ID
	args := c.Args()
	if len(args) < 2 {
		return c.Send(tr(uid, "ffp_usage"))
	}
	hash := resolveHash(c, args[0])
	if hash == "" {
		return c.Send(tr(uid, "invalid_hash"))
	}
	id, err := strconv.Atoi(args[1])
	if err != nil || id < 1 {
		return c.Send(tr(uid, "ffp_file_index"))
	}

	asJSON := false
	if len(args) >= 3 {
		last := strings.ToLower(strings.TrimSpace(args[len(args)-1]))
		if last == "json" || last == "--json" || last == "-j" {
			asJSON = true
		}
	}

	t := torr.GetTorrent(hash)
	if t == nil {
		return c.Send(tr(uid, "torrent_not_found"))
	}

	proto := "http"
	port := settings.Port
	if settings.Ssl {
		proto = "https"
		port = settings.SslPort
	}
	link := fmt.Sprintf("%s://127.0.0.1:%s/play/%s/%d", proto, port, hash, id)

	data, err := ffprobe.ProbeUrl(link)
	if err != nil {
		return c.Send(fmt.Sprintf(tr(uid, "ffp_error"), err.Error()))
	}

	var msg string
	if asJSON {
		buf, _ := json.MarshalIndent(data, "", "  ")
		msg = "<pre>" + strings.ReplaceAll(string(buf), "<", "&lt;") + "</pre>"
		if len(msg) > 4000 {
			msg = msg[:4000] + "\n...</pre>"
		}
	} else {
		msg = formatFfpHuman(data, uid)
		if len(msg) > 4000 {
			msg = msg[:4000] + "\n..."
		}
	}
	return c.Send(msg)
}

func formatFfpHuman(data *ffp.ProbeData, uid int64) string {
	var sb strings.Builder

	if data.Format != nil {
		f := data.Format
		sb.WriteString("<b>📁 " + tr(uid, "ffp_format") + "</b>\n")
		fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_container"), f.FormatLongName)
		if f.DurationSeconds > 0 {
			d := int(f.DurationSeconds)
			h, m, s := d/3600, (d%3600)/60, d%60
			fmt.Fprintf(&sb, "  %s: %02d:%02d:%02d\n", tr(uid, "ffp_duration"), h, m, s)
		}
		if f.Size != "" {
			if size, err := strconv.ParseInt(f.Size, 10, 64); err == nil {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_size"), humanize.IBytes(uint64(size)))
			} else {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_size"), f.Size)
			}
		}
		if f.BitRate != "" {
			if br, err := strconv.ParseInt(f.BitRate, 10, 64); err == nil {
				fmt.Fprintf(&sb, "  %s: %s/s\n", tr(uid, "ffp_bitrate"), humanize.IBytes(uint64(br)))
			} else {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_bitrate"), f.BitRate)
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("<b>🎬 " + tr(uid, "ffp_streams") + "</b>\n\n")
	for i, s := range data.Streams {
		title := getTag(s.TagList, "title")
		lang := getTag(s.TagList, "language")
		if lang != "" {
			lang = " [" + lang + "]"
		}

		switch s.CodecType {
		case "video":
			fmt.Fprintf(&sb, "<b>#%d %s</b>%s\n", i, tr(uid, "ffp_video"), lang)
			fmt.Fprintf(&sb, "  %s: %s", tr(uid, "ffp_codec"), s.CodecLongName)
			if s.Profile != "" {
				fmt.Fprintf(&sb, " (%s)", s.Profile)
			}
			sb.WriteString("\n")
			if s.Width > 0 && s.Height > 0 {
				fmt.Fprintf(&sb, "  %s: %d×%d", tr(uid, "ffp_resolution"), s.Width, s.Height)
				if s.DisplayAspectRatio != "" && s.DisplayAspectRatio != "0:0" {
					fmt.Fprintf(&sb, " (%s)", s.DisplayAspectRatio)
				}
				sb.WriteString("\n")
			}
			if s.PixFmt != "" {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_pixel"), s.PixFmt)
			}
			if s.RFrameRate != "" && s.RFrameRate != "0/0" {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_fps"), s.RFrameRate)
			}
			if s.BitRate != "" {
				if br, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil {
					fmt.Fprintf(&sb, "  %s: %s/s\n", tr(uid, "ffp_bitrate"), humanize.IBytes(uint64(br)))
				}
			}
			if s.ColorSpace != "" || s.ColorTransfer != "" {
				fmt.Fprintf(&sb, "  %s: %s / %s / %s\n", tr(uid, "ffp_color"), s.ColorSpace, s.ColorTransfer, s.ColorPrimaries)
			}
			if title != "" {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_title"), escapeHtml(title))
			}

		case "audio":
			fmt.Fprintf(&sb, "<b>#%d %s</b>%s\n", i, tr(uid, "ffp_audio"), lang)
			fmt.Fprintf(&sb, "  %s: %s", tr(uid, "ffp_codec"), s.CodecLongName)
			if s.Profile != "" {
				fmt.Fprintf(&sb, " (%s)", s.Profile)
			}
			sb.WriteString("\n")
			if s.SampleRate != "" {
				fmt.Fprintf(&sb, "  %s: %s Hz\n", tr(uid, "ffp_samplerate"), s.SampleRate)
			}
			if s.Channels > 0 {
				ch := s.ChannelLayout
				if ch == "" {
					ch = fmt.Sprintf("%d ch", s.Channels)
				}
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_channels"), ch)
			}
			if s.BitRate != "" {
				if br, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil {
					fmt.Fprintf(&sb, "  %s: %s/s\n", tr(uid, "ffp_bitrate"), humanize.IBytes(uint64(br)))
				}
			}
			if title != "" {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_title"), escapeHtml(title))
			}

		case "subtitle":
			fmt.Fprintf(&sb, "<b>#%d %s</b>%s\n", i, tr(uid, "ffp_subtitle"), lang)
			fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_codec"), s.CodecLongName)
			if title != "" {
				fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_title"), escapeHtml(title))
			}

		default:
			fmt.Fprintf(&sb, "<b>#%d %s</b>\n", i, s.CodecType)
			fmt.Fprintf(&sb, "  %s: %s\n", tr(uid, "ffp_codec"), s.CodecLongName)
		}
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n\n")
}

func getTag(tags ffp.Tags, key string) string {
	if tags == nil {
		return ""
	}
	if v, ok := tags[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	}
	for k, v := range tags {
		if strings.HasPrefix(k, key+"-") && v != nil {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprint(v)
		}
	}
	return ""
}
