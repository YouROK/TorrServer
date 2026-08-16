package bonjour

import "testing"

func TestSanitizeInstance(t *testing.T) {
	cases := map[string]string{
		"":                "TorrServer",
		"  ":              "TorrServer",
		"My Server":       "My Server",
		"My.Server.local": "My-Server",
		"My.Server.LOCAL": "My-Server",
		"Living Room TS":  "Living Room TS",
		"Room.local":      "Room",
	}
	for in, want := range cases {
		if got := sanitizeInstance(in); got != want {
			t.Fatalf("sanitizeInstance(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestStripLocalSuffix(t *testing.T) {
	cases := map[string]string{
		"MacBook-Pro-16---Pavel.local":  "MacBook-Pro-16---Pavel",
		"MacBook-Pro-16---Pavel.local.": "MacBook-Pro-16---Pavel",
		"host.LOCAL":                    "host",
		"plain":                         "plain",
	}
	for in, want := range cases {
		if got := stripLocalSuffix(in); got != want {
			t.Fatalf("stripLocalSuffix(%q)=%q, want %q", in, got, want)
		}
	}
}
