//go:build !windows

package offline

import (
	"golang.org/x/sys/unix"
)

// FreeSpace returns the number of bytes available to an unprivileged user on
// the filesystem that contains dir (Bavail, not Bfree: space reserved for the
// root user must not be counted as usable). The directory must exist: the
// callers create the offline directory before the check (checkOfflineDir).
// x/sys/unix is used instead of syscall so that the file compiles on every
// unix-like build target of the server, including android and darwin.
func FreeSpace(dir string) (int64, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return 0, err
	}
	return int64(st.Bavail) * int64(st.Bsize), nil
}
