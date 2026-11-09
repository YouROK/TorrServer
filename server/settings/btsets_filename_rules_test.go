package settings

import (
	"strings"
	"testing"
)

func TestSanitizeFileNameRules(t *testing.T) {
	in := []FileNameRule{
		{Name: "ok", Enabled: true, Mode: "extract", Pattern: `^(.+?) \d+`, Replacement: "x", Flags: "iig!"},
		{Name: "no pattern", Enabled: true, Mode: "replace"},
		{Name: "bad mode", Enabled: true, Mode: "weird", Pattern: "a", Replacement: "b"},
		{Name: "too long", Enabled: true, Pattern: strings.Repeat("a", maxFileNameRuleText+1)},
	}
	out := sanitizeFileNameRules(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(out))
	}
	if out[0].Mode != FileNameRuleExtract || out[0].Replacement != "" || out[0].Flags != "ig" {
		t.Fatalf("unexpected first rule: %+v", out[0])
	}
	if out[1].Mode != FileNameRuleReplace || out[1].Replacement != "b" {
		t.Fatalf("unexpected second rule: %+v", out[1])
	}
}

func TestSanitizeFileNameRulesNilAndLimit(t *testing.T) {
	if out := sanitizeFileNameRules(nil); out == nil || len(out) != 0 {
		t.Fatalf("nil input must give an empty non-nil slice, got %#v", out)
	}
	in := make([]FileNameRule, maxFileNameRules+10)
	for i := range in {
		in[i] = FileNameRule{Pattern: "a"}
	}
	if got := len(sanitizeFileNameRules(in)); got != maxFileNameRules {
		t.Fatalf("expected %d rules, got %d", maxFileNameRules, got)
	}
}

func TestSanitizeFileNameTemplate(t *testing.T) {
	if got := sanitizeFileNameTemplate("  {title} \n"); got != "{title}" {
		t.Fatalf("unexpected template: %q", got)
	}
	long := strings.Repeat("я", maxFileNameTemplate+10)
	if got := sanitizeFileNameTemplate(long); len([]rune(got)) != maxFileNameTemplate {
		t.Fatalf("template must be cut to %d runes, got %d", maxFileNameTemplate, len([]rune(got)))
	}
}

func TestSanitizeFileNameRulesCapture(t *testing.T) {
	out := sanitizeFileNameRules([]FileNameRule{
		{Name: "capture", Enabled: true, Mode: FileNameRuleCapture, Pattern: `(?<resolution>\d{3,4}p)`, Replacement: "x"},
	})
	if len(out) != 1 || out[0].Mode != FileNameRuleCapture || out[0].Replacement != "" {
		t.Fatalf("capture rule must be kept without replacement: %+v", out)
	}
}
