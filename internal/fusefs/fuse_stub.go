//go:build !(linux && !android) && !darwin && !freebsd

package fusefs

// platformSupported сообщает о недоступности FUSE на платформе.
func platformSupported() bool { return false }

// mountAt не монтирует: на платформе нет FUSE.
func mountAt(dir string, src FS, opts Options) (mountServer, error) {
	return nil, ErrUnsupported
}
