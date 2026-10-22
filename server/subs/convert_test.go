package subs

import (
	"strings"
	"testing"
)

func TestSRTToVTT(t *testing.T) {
	in := "1\n00:00:01,000 --> 00:00:02,500\nHello\n"
	want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:02.500\nHello\n"
	got, err := ToVTT([]byte(in), "srt")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func TestASSToVTT(t *testing.T) {
	in := `[Script Info]
Title: x

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:01.00,0:00:02.50,Default,,0,0,0,,{\i1}Hello, world{\i0}\Nsecond
Comment: 0,0:00:03.00,0:00:04.00,Default,,0,0,0,,skip
`
	want := "WEBVTT\n\n00:00:01.000 --> 00:00:02.500\nHello, world\nsecond\n\n"
	got, err := ToVTT([]byte(in), "ass")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("\nwant %q\ngot  %q", want, got)
	}
}

func TestASSWithoutDialogue(t *testing.T) {
	if _, err := ToVTT([]byte("[Events]\nFormat: Start, End, Text\n"), "ass"); err == nil {
		t.Error("expected error for ASS without dialogue lines")
	}
}

func TestCP1251(t *testing.T) {
	// "Привет" in Windows-1251
	cp := []byte{0xCF, 0xF0, 0xE8, 0xE2, 0xE5, 0xF2}
	in := append([]byte("1\n00:00:01,000 --> 00:00:02,000\n"), cp...)
	in = append(in, '\n')
	got, err := ToVTT(in, "srt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Привет") {
		t.Errorf("cp1251 text not decoded: %q", got)
	}
}

func TestCP1251Ukrainian(t *testing.T) {
	// "Їжак і єнот, ґанок" in Windows-1251
	cp := []byte{0xAF, 0xE6, 0xE0, 0xEA, 0x20, 0xB3, 0x20, 0xBA, 0xED, 0xEE, 0xF2, 0x2C, 0x20, 0xB4, 0xE0, 0xED, 0xEE, 0xEA}
	if got := Decode(cp); got != "Їжак і єнот, ґанок" {
		t.Errorf("decode: %q", got)
	}
	if b, _, _ := Encode("Їжак і єнот, ґанок", "cp1251"); string(b) != string(cp) {
		t.Errorf("encode: %x", b)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	if _, err := ToVTT([]byte("x"), "sub"); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestUTF8WithStrayByte(t *testing.T) {
	in := []byte("1\n00:00:01,000 --> 00:00:02,000\nПривет, мир")
	in = append(in, 0xFF) // a single stray byte
	in = append(in, '\n')
	got, err := ToVTT(in, "srt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Привет, мир") {
		t.Errorf("UTF-8 text was corrupted: %q", got)
	}
}

func TestEncode(t *testing.T) {
	// UTF-8 by default
	b, cs, err := Encode("Привет", "")
	if err != nil || cs != "utf-8" || string(b) != "Привет" {
		t.Errorf("default: %q %q %v", b, cs, err)
	}

	// Windows-1251 (and an alias)
	want := []byte{0xCF, 0xF0, 0xE8, 0xE2, 0xE5, 0xF2}
	for _, name := range []string{"windows-1251", "CP1251"} {
		b, cs, err = Encode("Привет", name)
		if err != nil || cs != "windows-1251" || string(b) != string(want) {
			t.Errorf("%s: %x %q %v", name, b, cs, err)
		}
	}

	// Round trip: cp1251 -> UTF-8
	if got := Decode(want); got != "Привет" {
		t.Errorf("round trip: %q", got)
	}

	// A character outside the encoding is replaced with '?'
	b, _, _ = Encode("a→b", "cp1251")
	if string(b) != "a?b" {
		t.Errorf("unrepresentable rune: %q", b)
	}

	// Unsupported encoding
	if _, _, err = Encode("x", "koi8-r"); err == nil {
		t.Error("expected error for unsupported encoding")
	}
}
