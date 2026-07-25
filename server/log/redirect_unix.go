//go:build !windows

package log

import (
	"os"

	"golang.org/x/sys/unix"
)

func redirectStdFDs(f *os.File) error {
	fd := int(f.Fd())
	if err := unix.Dup2(fd, 1); err != nil {
		return err
	}
	if err := unix.Dup2(fd, 2); err != nil {
		return err
	}
	os.Stdout = f
	os.Stderr = f
	return nil
}
