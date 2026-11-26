package api

import "testing"

func TestUnescapeQuery(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Plain Title", "Plain Title"},
		{"Caf%C3%A9", "Café"},
		{"100% Pure", "100% Pure"},
		{"50%", "50%"},
		{"%zz", "%zz"},
		{"", ""},
	} {
		if got := unescapeQuery(tc.in); got != tc.want {
			t.Errorf("unescapeQuery(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
