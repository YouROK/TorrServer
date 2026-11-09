package subs

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var (
	srtTime  = regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`)
	assTags  = regexp.MustCompile(`\{[^}]*\}`)
	assBreak = strings.NewReplacer(`\N`, "\n", `\n`, "\n", `\h`, " ")
	vttEsc   = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

// ToVTT converts srt/ass/ssa/vtt subtitles to WebVTT.
// ASS styling (colors, positions, fonts) is dropped; only the text is kept.
func ToVTT(raw []byte, format string) (string, error) {
	text := decode(raw)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	switch strings.ToLower(strings.TrimPrefix(format, ".")) {
	case "vtt":
		if strings.HasPrefix(text, "WEBVTT") {
			return text, nil
		}
		return "WEBVTT\n\n" + text, nil
	case "srt":
		return "WEBVTT\n\n" + srtTime.ReplaceAllString(text, "$1.$2"), nil
	case "ass", "ssa":
		return assToVTT(text)
	}
	return "", fmt.Errorf("unsupported subtitle format %q", format)
}

type cue struct {
	start, end int // milliseconds
	text       string
}

func assToVTT(text string) (string, error) {
	var (
		inEvents bool
		fields   []string
		iStart   = -1
		iEnd     = -1
		iText    = -1
		cues     []cue
	)

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "[") {
			inEvents = strings.EqualFold(line, "[Events]")
			continue
		}
		if !inEvents {
			continue
		}

		switch {
		case strings.HasPrefix(line, "Format:"):
			fields = splitTrim(strings.TrimPrefix(line, "Format:"))
			for i, f := range fields {
				switch strings.ToLower(f) {
				case "start":
					iStart = i
				case "end":
					iEnd = i
				case "text":
					iText = i
				}
			}

		case strings.HasPrefix(line, "Dialogue:"):
			if iStart < 0 || iEnd < 0 || iText < 0 {
				continue
			}
			// Text is the last field and may contain commas.
			parts := strings.SplitN(strings.TrimPrefix(line, "Dialogue:"), ",", len(fields))
			if len(parts) < len(fields) {
				continue
			}
			s, ok1 := parseASSTime(parts[iStart])
			e, ok2 := parseASSTime(parts[iEnd])
			t := cleanASSText(parts[iText])
			if !ok1 || !ok2 || e <= s || t == "" {
				continue
			}
			cues = append(cues, cue{s, e, t})
		}
	}

	if len(cues) == 0 {
		return "", errors.New("no dialogue lines found")
	}
	sort.SliceStable(cues, func(i, j int) bool { return cues[i].start < cues[j].start })

	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	for _, c := range cues {
		fmt.Fprintf(&sb, "%s --> %s\n%s\n\n", fmtTime(c.start), fmtTime(c.end), c.text)
	}
	return sb.String(), nil
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// "0:01:23.45" -> milliseconds
func parseASSTime(s string) (int, bool) {
	var h, m, sec, cs int
	n, err := fmt.Sscanf(strings.TrimSpace(s), "%d:%d:%d.%d", &h, &m, &sec, &cs)
	if err != nil || n != 4 {
		return 0, false
	}
	return ((h*60+m)*60+sec)*1000 + cs*10, true
}

func fmtTime(ms int) string {
	return fmt.Sprintf("%02d:%02d:%02d.%03d", ms/3600000, ms/60000%60, ms/1000%60, ms%1000)
}

func cleanASSText(s string) string {
	if strings.Contains(s, `\p1`) { // vector drawing (karaoke/signs), not text
		return ""
	}
	s = assTags.ReplaceAllString(s, "")
	s = vttEsc.Replace(s)
	s = assBreak.Replace(s)
	for strings.Contains(s, "\n\n") { // a blank line terminates a cue in VTT
		s = strings.ReplaceAll(s, "\n\n", "\n")
	}
	return strings.TrimSpace(s)
}

// --- input encoding --------------------------------------------------------

// Decode converts subtitle bytes to UTF-8: it strips the BOM and falls back to
// Windows-1251 when the data is not UTF-8 (Russian subtitles are often cp1251).
// UTF-16 is not handled. If golang.org/x/text is already a dependency,
// charmap.Windows1251 can replace the hand-written table below.
func Decode(b []byte) string { return decode(b) }

func decode(b []byte) string {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(b) {
		return string(b)
	}
	// The file is almost entirely UTF-8 with a few stray bytes: do not mistake it
	// for cp1251, which would turn the whole text into mojibake.
	if mostlyUTF8(b) {
		return strings.ToValidUTF8(string(b), "\uFFFD")
	}
	return cp1251ToUTF8(b)
}

// mostlyUTF8 reports whether the vast majority of bytes >= 0x80 form valid
// multi-byte UTF-8 sequences (up to ~10% bad bytes are tolerated).
// Real cp1251 text almost never produces such sequences.
func mostlyUTF8(b []byte) bool {
	good, bad := 0, 0
	for i := 0; i < len(b); {
		if b[i] < 0x80 {
			i++
			continue
		}
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && size == 1 {
			bad++
			i++
			continue
		}
		good += size
		i += size
	}
	return good > 0 && bad*10 <= good
}

// cp1251High maps bytes 0x80..0xBF; 0 marks the one unassigned byte (0x98).
var cp1251High = [64]rune{
	'Ђ', 'Ѓ', '‚', 'ѓ', '„', '…', '†', '‡', '€', '‰', 'Љ', '‹', 'Њ', 'Ќ', 'Ћ', 'Џ', // 0x8_
	'ђ', '‘', '’', '“', '”', '•', '–', '—', 0, '™', 'љ', '›', 'њ', 'ќ', 'ћ', 'џ', // 0x9_
	'\u00A0', 'Ў', 'ў', 'Ј', '¤', 'Ґ', '¦', '§', 'Ё', '©', 'Є', '«', '¬', '\u00AD', '®', 'Ї', // 0xA_
	'°', '±', 'І', 'і', 'ґ', 'µ', '¶', '·', 'ё', '№', 'є', '»', 'ј', 'Ѕ', 'ѕ', 'ї', // 0xB_
}

func cp1251ToUTF8(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		switch {
		case c < 0x80:
			sb.WriteByte(c)
		case c >= 0xC0: // Cyrillic letters (U+0410..U+044F)
			sb.WriteRune(rune(c) - 0xC0 + 0x0410)
		case cp1251High[c-0x80] != 0:
			sb.WriteRune(cp1251High[c-0x80])
		default:
			sb.WriteRune('\uFFFD')
		}
	}
	return sb.String()
}

// --- output encoding -------------------------------------------------------

// Encode converts UTF-8 text to the requested encoding. It returns the bytes and
// the canonical encoding name for the Content-Type header.
// Supported: "utf8"/"utf-8" (the default, also for an empty string) and
// "windows-1251" (aliases: "cp1251", "win1251", "1251").
// Characters missing in the target encoding are replaced with '?'.
func Encode(text, encoding string) ([]byte, string, error) {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(encoding), "_", "-")) {
	case "", "utf8", "utf-8":
		return []byte(text), "utf-8", nil
	case "windows-1251", "cp1251", "win1251", "1251":
		return utf8ToCP1251(text), "windows-1251", nil
	}
	return nil, "", fmt.Errorf("unsupported encoding %q (supported: utf8, windows-1251)", encoding)
}

var cp1251Reverse = func() map[rune]byte {
	m := make(map[rune]byte, 128)
	for c := 0xC0; c <= 0xFF; c++ { // Cyrillic letters (U+0410..U+044F)
		m[rune(c)-0xC0+0x0410] = byte(c)
	}
	for i, r := range cp1251High {
		if r != 0 {
			m[r] = byte(0x80 + i)
		}
	}
	return m
}()

func utf8ToCP1251(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r < 0x80 {
			out = append(out, byte(r))
			continue
		}
		if c, ok := cp1251Reverse[r]; ok {
			out = append(out, c)
		} else {
			out = append(out, '?')
		}
	}
	return out
}
