package torrfs

func New() *RootDir {
	r := NewRootDir()
	r.buildChildren()
	return r
}
