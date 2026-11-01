//go:build windows

package offline

import (
	"golang.org/x/sys/windows"
)

// FreeSpace returns the number of bytes available to the current user on the
// disk that contains dir (freeBytesAvailableToCaller, i.e. already adjusted
// for quotas). The directory must exist: the callers create the offline
// directory before the check (checkOfflineDir).
func FreeSpace(dir string) (int64, error) {
	p, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return 0, err
	}
	var avail, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(p, &avail, &total, &free); err != nil {
		return 0, err
	}
	return int64(avail), nil
}
