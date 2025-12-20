package webdav

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"server/log"
	"sync"
	"time"

	"server/torrfs"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

var missingMethods = []string{
	"PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK",
}

func MountWebDAV(r *gin.Engine) {
	log.TLogln("Starting WebDAV")
	tfs := torrfs.AsFS(torrfs.New())

	h := &webdav.Handler{
		Prefix:     "/dav",
		FileSystem: &ReadOnlyFS{FS: tfs},
		LockSystem: webdav.NewMemLS(),
	}

	grp := r.Group("/dav")

	handler := func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}

	grp.Any("/*webdav", handler)
	for _, m := range missingMethods {
		grp.Handle(m, "/*webdav", handler)
	}

	grp.Any("", handler)
	for _, m := range missingMethods {
		grp.Handle(m, "", handler)
	}
}

type ReadOnlyFS struct {
	FS fs.FS
}

var _ webdav.FileSystem = (*ReadOnlyFS)(nil)

func (ro *ReadOnlyFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return os.ErrPermission
}

func (ro *ReadOnlyFS) RemoveAll(ctx context.Context, name string) error {
	return os.ErrPermission
}

func (ro *ReadOnlyFS) Rename(ctx context.Context, oldName, newName string) error {
	return os.ErrPermission
}

func (ro *ReadOnlyFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = cleanWebDAVPath(name)
	return fs.Stat(ro.FS, name)
}

func (ro *ReadOnlyFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, os.ErrPermission
	}

	name = cleanWebDAVPath(name)

	f, err := ro.FS.Open(name)
	if err != nil {
		return nil, err
	}

	return newROFile(ro.FS, name, f), nil
}

// --- file wrapper ---

type roFile struct {
	fsys fs.FS
	name string

	mu sync.Mutex
	f  fs.File

	dirPos  int
	dirList []fs.DirEntry
}

func newROFile(fsys fs.FS, name string, f fs.File) *roFile {
	return &roFile{fsys: fsys, name: name, f: f}
}

var _ webdav.File = (*roFile)(nil)

func (f *roFile) Write(p []byte) (n int, err error) {
	return 0, fs.ErrPermission
}

func (f *roFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.f == nil {
		return nil
	}
	err := f.f.Close()
	f.f = nil
	f.dirList = nil
	f.dirPos = 0
	return err
}

func (f *roFile) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.f == nil {
		return 0, fs.ErrClosed
	}
	r, ok := f.f.(io.Reader)
	if !ok {
		return 0, fs.ErrInvalid
	}
	return r.Read(p)
}

func (f *roFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.f == nil {
		return 0, fs.ErrClosed
	}
	rs, ok := f.f.(io.Seeker)
	if !ok {
		return 0, errors.New("seek not supported")
	}
	return rs.Seek(offset, whence)
}

func (f *roFile) Stat() (os.FileInfo, error) {
	return fs.Stat(f.fsys, f.name)
}

func (f *roFile) Readdir(count int) ([]os.FileInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.f == nil {
		return nil, fs.ErrClosed
	}

	fi, err := fs.Stat(f.fsys, f.name)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fs.ErrInvalid
	}

	if f.dirList == nil {
		des, err := fs.ReadDir(f.fsys, f.name)
		if err != nil {
			return nil, err
		}
		f.dirList = des
		f.dirPos = 0
	}

	if count <= 0 {
		out := make([]os.FileInfo, 0, len(f.dirList)-f.dirPos)
		for f.dirPos < len(f.dirList) {
			de := f.dirList[f.dirPos]
			f.dirPos++
			info, err := de.Info()
			if err != nil {
				continue
			}
			out = append(out, info)
		}
		return out, nil
	}

	out := make([]os.FileInfo, 0, count)
	for f.dirPos < len(f.dirList) && len(out) < count {
		de := f.dirList[f.dirPos]
		f.dirPos++
		info, err := de.Info()
		if err != nil {
			continue
		}
		out = append(out, info)
	}

	if len(out) == 0 {
		return nil, io.EOF
	}
	return out, nil
}

// --- path helpers ---
func cleanWebDAVPath(name string) string {
	if name == "" || name == "/" {
		return "."
	}
	name = path.Clean("/" + name)
	name = name[1:]
	if name == "" {
		return "."
	}
	return name
}

func nonZeroTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Unix(0, 0)
	}
	return t
}
