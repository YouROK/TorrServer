//go:build (linux && !android) || darwin || freebsd

package fusefs

import (
	"context"
	"errors"
	"hash/fnv"
	"io"
	"os"
	"path"
	"sync"
	"syscall"
	"time"

	"silo/internal/log"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	dirPerm  = 0555
	filePerm = 0444
)

func platformSupported() bool { return true }

// mountAt монтирует дерево источника в каталог.
func mountAt(dir string, src FS, opts Options) (mountServer, error) {
	if err := preflightMount(opts); err != nil {
		return nil, err
	}

	files := newOpenFiles()
	root := &node{src: src, open: files}

	oneSec := time.Second
	o := &fs.Options{
		MountOptions: fuse.MountOptions{
			Name:       mountName,
			FsName:     mountFsName,
			AllowOther: opts.AllowOther,
			Options:    []string{"ro"},
		},
		EntryTimeout:    &oneSec,
		AttrTimeout:     &oneSec,
		NegativeTimeout: &oneSec,
	}

	srv, err := fs.Mount(dir, root, o)
	if err != nil {
		return nil, err
	}

	log.Infof("[Fuse] Mounted %s at %q", mountFsName, dir)
	return newMountedServer(srv, files), nil
}

// mountedServer закрывает открытые файлы перед размонтированием: чтение,
// ожидающее кусок торрента, иначе задержит снятие файловой системы.
type mountedServer struct {
	server *fuse.Server
	files  *openFiles
	stop   chan struct{}
	once   sync.Once
}

func newMountedServer(srv *fuse.Server, files *openFiles) *mountedServer {
	return &mountedServer{
		server: srv,
		files:  files,
		stop:   make(chan struct{}),
	}
}

func (m *mountedServer) Unmount() error {
	m.files.closeAll()
	return m.server.Unmount()
}

// done закрывается, когда сервер прекратил обслуживать запросы.
// Это происходит и при снятии монтирования извне, мимо Unmount.
func (m *mountedServer) done() <-chan struct{} {
	m.once.Do(func() {
		go func() {
			m.server.Wait()
			close(m.stop)
		}()
	})
	return m.stop
}

// openFiles хранит открытые файлы смонтированной файловой системы.
type openFiles struct {
	mu    sync.Mutex
	items map[*fileHandle]struct{}
}

func newOpenFiles() *openFiles {
	return &openFiles{items: make(map[*fileHandle]struct{})}
}

func (o *openFiles) add(h *fileHandle) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.items[h] = struct{}{}
}

func (o *openFiles) remove(h *fileHandle) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.items, h)
}

func (o *openFiles) closeAll() {
	o.mu.Lock()
	handles := make([]*fileHandle, 0, len(o.items))
	for h := range o.items {
		handles = append(handles, h)
	}
	o.items = make(map[*fileHandle]struct{})
	o.mu.Unlock()

	for _, h := range handles {
		h.close()
	}
}

// node - узел файловой системы: каталог или файл.
// Метаданные берутся из источника при создании узла, а содержимое
// каталога запрашивается заново на каждый readdir.
type node struct {
	fs.Inode

	src  FS
	open *openFiles
	info Node   // nil только у корня
	path string // путь внутри источника, у корня пустой
}

// newNode создаёт узел по описанию из источника.
func newNode(src FS, info Node, p string) *node {
	return &node{src: src, info: info, path: p}
}

func (n *node) isDir() bool {
	return n.info == nil || n.info.IsDir()
}

var (
	_ = (fs.NodeLookuper)((*node)(nil))
	_ = (fs.NodeReaddirer)((*node)(nil))
	_ = (fs.NodeGetattrer)((*node)(nil))
	_ = (fs.NodeOpener)((*node)(nil))
)

// Lookup отдаёт узел по имени внутри каталога.
func (n *node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if !n.isDir() {
		return nil, syscall.ENOTDIR
	}

	child, err := n.src.Stat(joinPath(n.path, name))
	if err != nil {
		return nil, errnoFor(err)
	}

	childPath := joinPath(n.path, name)
	childNode := newNode(n.src, child, childPath)
	childNode.open = n.open

	ch := n.NewInode(ctx, childNode, stableAttr(childPath, child.IsDir()))
	fillAttr(&out.Attr, child)
	return ch, 0
}

// Readdir перечисляет содержимое каталога.
func (n *node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	if !n.isDir() {
		return nil, syscall.ENOTDIR
	}

	children, err := n.src.List(n.path)
	if err != nil {
		return nil, errnoFor(err)
	}

	entries := make([]fuse.DirEntry, 0, len(children))
	for _, c := range children {
		p := joinPath(n.path, c.Name())
		entries = append(entries, fuse.DirEntry{
			Name: c.Name(),
			Mode: modeFor(c.IsDir()),
			Ino:  stableAttr(p, c.IsDir()).Ino,
		})
	}
	return fs.NewListDirStream(entries), 0
}

// Getattr отдаёт атрибуты узла.
func (n *node) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	if n.info == nil {
		fillDirAttr(&out.Attr)
		return 0
	}
	fillAttr(&out.Attr, n.info)
	return 0
}

// Open открывает файл на чтение.
func (n *node) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	if n.isDir() {
		return nil, 0, syscall.EISDIR
	}
	if flags&(syscall.O_WRONLY|syscall.O_RDWR|syscall.O_TRUNC|syscall.O_APPEND) != 0 {
		return nil, 0, syscall.EROFS
	}

	f, err := n.src.Open(n.path)
	if err != nil {
		return nil, 0, errnoFor(err)
	}

	h := &fileHandle{f: f, open: n.open}
	if n.open != nil {
		n.open.add(h)
	}
	return h, 0, 0
}

// fileHandle читает открытый файл по смещению.
// Чтения одного дескриптора сериализуются: источник допускает один Seek на файл.
type fileHandle struct {
	mu   sync.Mutex
	f    File
	open *openFiles
	once sync.Once
}

var (
	_ = (fs.FileReader)((*fileHandle)(nil))
	_ = (fs.FileReleaser)((*fileHandle)(nil))
)

func (h *fileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := h.f.Seek(off, io.SeekStart); err != nil {
		return nil, errnoFor(err)
	}

	read, err := io.ReadFull(h.f, dest)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, errnoFor(err)
	}
	return fuse.ReadResultData(dest[:read]), 0
}

func (h *fileHandle) Release(ctx context.Context) syscall.Errno {
	h.close()
	if h.open != nil {
		h.open.remove(h)
	}
	return 0
}

// close закрывает файл; повторный вызов ничего не делает.
func (h *fileHandle) close() {
	h.once.Do(func() {
		if err := h.f.Close(); err != nil {
			log.Warnf("[Fuse] Failed to close file handle: %v", err)
		}
	})
}

// fillAttr переносит метаданные узла в атрибуты файловой системы.
func fillAttr(out *fuse.Attr, n Node) {
	if n.IsDir() {
		fillDirAttr(out)
		return
	}

	out.Mode = modeFor(false)
	out.Size = uint64(n.Size())
	out.Nlink = 1
	out.Blksize = 4096
	out.Blocks = (out.Size + 511) / 512
	out.Mtime, out.Mtimensec = splitTime(n.ModTime())
	out.Atime, out.Atimensec = out.Mtime, out.Mtimensec
	out.Ctime, out.Ctimensec = out.Mtime, out.Mtimensec
}

func fillDirAttr(out *fuse.Attr) {
	out.Mode = modeFor(true)
	out.Nlink = 2
	out.Blksize = 4096
	out.Mtime, out.Mtimensec = splitTime(time.Time{})
	out.Atime, out.Atimensec = out.Mtime, out.Mtimensec
	out.Ctime, out.Ctimensec = out.Mtime, out.Mtimensec
}

func modeFor(isDir bool) uint32 {
	if isDir {
		return syscall.S_IFDIR | dirPerm
	}
	return syscall.S_IFREG | filePerm
}

// splitTime делит время на секунды и наносекунды, приводя нулевое время к началу эпохи.
func splitTime(t time.Time) (uint64, uint32) {
	if t.IsZero() {
		return 0, 0
	}
	t = t.UTC()
	return uint64(t.Unix()), uint32(t.Nanosecond())
}

// stableAttr даёт постоянный inode для пути: обращения к одному и тому же
// пути должны приводить к одному узлу дерева.
func stableAttr(p string, isDir bool) fs.StableAttr {
	h := fnv.New64a()
	_, _ = h.Write([]byte(p))
	if isDir {
		_, _ = h.Write([]byte{1})
	}

	ino := h.Sum64()
	if ino == 0 || ino == ^uint64(0) {
		ino = 1
	}
	return fs.StableAttr{Mode: modeFor(isDir), Ino: ino}
}

// errnoFor переводит ошибку источника в код файловой системы.
func errnoFor(err error) syscall.Errno {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return syscall.ENOENT
	case errors.Is(err, os.ErrPermission):
		return syscall.EACCES
	default:
		return syscall.EIO
	}
}

func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return path.Join(parent, name)
}
