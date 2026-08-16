//go:build windows

package log

import "os"

func redirectStdFDs(f *os.File) error {
	os.Stdout = f
	os.Stderr = f
	return nil
}
