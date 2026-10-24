//go:build windows

package log

import "testing"

func withRestoredStdio(t *testing.T) { t.Helper() }
