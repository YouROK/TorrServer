//go:build darwin || freebsd

package fusefs

// preflightMount не требует проверок: на этих системах allow_other
// не зависит от fuse.conf.
func preflightMount(opts Options) error {
	return nil
}
