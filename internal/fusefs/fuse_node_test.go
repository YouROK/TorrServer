//go:build (linux && !android) || darwin || freebsd

package fusefs

import (
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// fakeNode - узел поддельного источника.
type fakeNode struct {
	name  string
	isDir bool
	size  int64
}

func (n fakeNode) Name() string       { return n.name }
func (n fakeNode) IsDir() bool        { return n.isDir }
func (n fakeNode) Size() int64        { return n.size }
func (n fakeNode) ModTime() time.Time { return time.Unix(1700000000, 0) }

// fakeFile - файл поддельного источника.
type fakeFile struct {
	*strings.Reader
}

func (f fakeFile) Name() string       { return "file.mkv" }
func (f fakeFile) ModTime() time.Time { return time.Unix(1700000000, 0) }
func (f fakeFile) Close() error       { return nil }

// fakeFS отдаёт фиксированное дерево.
type fakeFS struct{}

func (fakeFS) Stat(p string) (Node, error) {
	switch p {
	case "":
		return fakeNode{isDir: true}, nil
	case "movie":
		return fakeNode{name: "movie", isDir: true}, nil
	case "movie/file.mkv":
		return fakeNode{name: "file.mkv", size: 2048}, nil
	default:
		return nil, os.ErrNotExist
	}
}

func (fakeFS) List(p string) ([]Node, error) {
	switch p {
	case "":
		return []Node{fakeNode{name: "movie", isDir: true}}, nil
	case "movie":
		return []Node{fakeNode{name: "file.mkv", size: 2048}}, nil
	default:
		return nil, os.ErrNotExist
	}
}

func (fakeFS) Open(p string) (File, error) {
	if p != "movie/file.mkv" {
		return nil, os.ErrNotExist
	}
	return fakeFile{strings.NewReader("0123456789")}, nil
}

// newBridge собирает файловую систему без монтирования: мост позволяет
// вызывать Lookup так же, как это делает ядро.
func newBridge(t *testing.T) fuse.RawFileSystem {
	t.Helper()

	root := newNode(fakeFS{}, nil, "")
	return fs.NewNodeFS(root, &fs.Options{})
}

func TestRootNodeIsDir(t *testing.T) {
	root := newNode(fakeFS{}, nil, "")

	if !root.isDir() {
		t.Error("корень не каталог")
	}
}

func TestLookupMissingPath(t *testing.T) {
	raw := newBridge(t)

	var out fuse.EntryOut
	if status := raw.Lookup(nil, &fuse.InHeader{NodeId: 1}, "nope", &out); status != fuse.ENOENT {
		t.Errorf("status = %v, want ENOENT", status)
	}
}

func TestLookupReturnsFileAttr(t *testing.T) {
	raw := newBridge(t)

	var out fuse.EntryOut
	if status := raw.Lookup(nil, &fuse.InHeader{NodeId: 1}, "movie", &out); status != fuse.OK {
		t.Fatalf("lookup movie: %v", status)
	}
	if out.Mode&syscall.S_IFDIR == 0 {
		t.Errorf("mode = %o, want каталог", out.Mode)
	}

	if status := raw.Lookup(nil, &fuse.InHeader{NodeId: out.NodeId}, "file.mkv", &out); status != fuse.OK {
		t.Fatalf("lookup file: %v", status)
	}
	if out.Size != 2048 {
		t.Errorf("size = %d, want 2048", out.Size)
	}
	if out.Mode&syscall.S_IFDIR != 0 {
		t.Errorf("mode = %o, want обычный файл", out.Mode)
	}
}

// TestStableAttrIsStable проверяет, что один путь даёт один inode:
// иначе ядро будет считать файл новым при каждом обращении.
func TestStableAttrIsStable(t *testing.T) {
	a := stableAttr("movie/file.mkv", false)
	b := stableAttr("movie/file.mkv", false)
	if a.Ino != b.Ino {
		t.Errorf("ino = %d и %d, want одинаковые", a.Ino, b.Ino)
	}

	c := stableAttr("movie/other.mkv", false)
	if a.Ino == c.Ino {
		t.Error("разные пути получили один inode")
	}

	d := stableAttr("movie/file.mkv", true)
	if a.Ino == d.Ino {
		t.Error("каталог и файл с одним путём получили один inode")
	}
}

// TestStableAttrAvoidsReserved проверяет, что inode не попадает в
// зарезервированное значение: go-fuse паникует на таком узле.
func TestStableAttrAvoidsReserved(t *testing.T) {
	for _, p := range []string{"", "a", "movie/file.mkv", "категория/раздача/файл"} {
		if got := stableAttr(p, false); got.Reserved() {
			t.Errorf("путь %q получил зарезервированный inode", p)
		}
	}
}

func TestReadAtOffset(t *testing.T) {
	n := newNode(fakeFS{}, fakeNode{name: "file.mkv", size: 10}, "movie/file.mkv")

	fh, _, errno := n.Open(nil, uint32(os.O_RDONLY))
	if errno != 0 {
		t.Fatalf("open errno = %v", errno)
	}
	defer fh.(fs.FileReleaser).Release(nil)

	h, ok := fh.(*fileHandle)
	if !ok {
		t.Fatalf("handle = %T, want *fileHandle", fh)
	}

	res, errno := h.Read(nil, make([]byte, 3), 4)
	if errno != 0 {
		t.Fatalf("read errno = %v", errno)
	}
	data, _ := res.Bytes(nil)
	if string(data) != "456" {
		t.Errorf("data = %q, want %q", data, "456")
	}
}

// TestReadPastEnd возвращает прочитанное без ошибки: короткое чтение
// в конце файла это норма, а не сбой.
func TestReadPastEnd(t *testing.T) {
	n := newNode(fakeFS{}, fakeNode{name: "file.mkv", size: 10}, "movie/file.mkv")

	fh, _, _ := n.Open(nil, uint32(os.O_RDONLY))
	h, ok := fh.(*fileHandle)
	if !ok {
		t.Fatalf("handle = %T, want *fileHandle", fh)
	}

	res, errno := h.Read(nil, make([]byte, 16), 8)
	if errno != 0 {
		t.Fatalf("read errno = %v", errno)
	}
	data, _ := res.Bytes(nil)
	if string(data) != "89" {
		t.Errorf("data = %q, want %q", data, "89")
	}
}

func TestOpenRejectsWrite(t *testing.T) {
	n := newNode(fakeFS{}, fakeNode{name: "file.mkv", size: 10}, "movie/file.mkv")

	if _, _, errno := n.Open(nil, uint32(os.O_WRONLY)); errno != syscall.EROFS {
		t.Errorf("errno = %v, want EROFS", errno)
	}
}

func TestOpenRejectsDir(t *testing.T) {
	dir := newNode(fakeFS{}, fakeNode{name: "movie", isDir: true}, "movie")

	if _, _, errno := dir.Open(nil, uint32(os.O_RDONLY)); errno != syscall.EISDIR {
		t.Errorf("errno = %v, want EISDIR", errno)
	}
}

// TestReaddirListsChildren проверяет состав выдачи каталога.
func TestReaddirListsChildren(t *testing.T) {
	dir := newNode(fakeFS{}, fakeNode{name: "movie", isDir: true}, "movie")

	ds, errno := dir.Readdir(nil)
	if errno != 0 {
		t.Fatalf("readdir errno = %v", errno)
	}
	defer ds.Close()

	var names []string
	for ds.HasNext() {
		e, errno := ds.Next()
		if errno != 0 {
			t.Fatalf("next errno = %v", errno)
		}
		names = append(names, e.Name)
	}

	if len(names) != 1 || names[0] != "file.mkv" {
		t.Errorf("names = %v, want [file.mkv]", names)
	}
}

func TestErrnoForNotExist(t *testing.T) {
	if got := errnoFor(os.ErrNotExist); got != syscall.ENOENT {
		t.Errorf("errno = %v, want ENOENT", got)
	}
	if got := errnoFor(os.ErrPermission); got != syscall.EACCES {
		t.Errorf("errno = %v, want EACCES", got)
	}
	if got := errnoFor(io.ErrUnexpectedEOF); got != syscall.EIO {
		t.Errorf("errno = %v, want EIO", got)
	}
}

func TestJoinPath(t *testing.T) {
	cases := []struct{ parent, name, want string }{
		{"", "movie", "movie"},
		{"movie", "file.mkv", "movie/file.mkv"},
		{"a/b", "c", "a/b/c"},
	}
	for _, c := range cases {
		if got := joinPath(c.parent, c.name); got != c.want {
			t.Errorf("joinPath(%q, %q) = %q, want %q", c.parent, c.name, got, c.want)
		}
	}
}

// countingFile считает закрытия: closeAll не должен закрывать файл дважды.
type countingFile struct {
	*strings.Reader
	closed int
}

func (f *countingFile) Name() string       { return "file.mkv" }
func (f *countingFile) ModTime() time.Time { return time.Unix(1700000000, 0) }

func (f *countingFile) Close() error {
	f.closed++
	return nil
}

// TestOpenFilesCloseAll проверяет, что перед размонтированием файлы
// закрываются: иначе чтение, ждущее кусок торрента, задержит снятие.
func TestOpenFilesCloseAll(t *testing.T) {
	open := newOpenFiles()
	f := &countingFile{Reader: strings.NewReader("data")}
	h := &fileHandle{f: f, open: open}
	open.add(h)

	open.closeAll()
	if f.closed != 1 {
		t.Errorf("закрытий = %d, want 1", f.closed)
	}

	open.closeAll()
	if f.closed != 1 {
		t.Errorf("повторный closeAll закрыл файл ещё раз: %d", f.closed)
	}
}

// TestOpenFilesRemove проверяет, что закрытый ядром дескриптор больше не
// отслеживается и не закрывается повторно при размонтировании.
func TestOpenFilesRemove(t *testing.T) {
	open := newOpenFiles()
	f := &countingFile{Reader: strings.NewReader("data")}
	h := &fileHandle{f: f, open: open}
	open.add(h)

	h.Release(nil)
	open.closeAll()

	if f.closed != 1 {
		t.Errorf("закрытий = %d, want 1", f.closed)
	}
}

// TestOpenFilesInheritedByChild проверяет, что узел файла получает список
// открытых дескрипторов от родителя: иначе размонтирование их не увидит.
func TestOpenFilesInheritedByChild(t *testing.T) {
	open := newOpenFiles()
	root := newNode(fakeFS{}, nil, "")
	root.open = open
	raw := fs.NewNodeFS(root, &fs.Options{})

	var dirOut fuse.EntryOut
	if status := raw.Lookup(nil, &fuse.InHeader{NodeId: 1}, "movie", &dirOut); status != fuse.OK {
		t.Fatalf("lookup movie: %v", status)
	}

	var fileOut fuse.EntryOut
	if status := raw.Lookup(nil, &fuse.InHeader{NodeId: dirOut.NodeId}, "file.mkv", &fileOut); status != fuse.OK {
		t.Fatalf("lookup file: %v", status)
	}

	// Проверяем, что узел файла, созданный мостом, унаследовал список
	// открытых дескрипторов: иначе размонтирование их не увидит.
	fileNode := root.GetChild("movie")
	if fileNode == nil {
		t.Fatal("узел movie не найден в дереве")
	}

	fileChild := fileNode.GetChild("file.mkv")
	if fileChild == nil {
		t.Fatal("узел file.mkv не найден в дереве")
	}

	ops, ok := fileChild.Operations().(*node)
	if !ok {
		t.Fatalf("операции = %T, want *node", fileChild.Operations())
	}
	if ops.open != open {
		t.Error("дочерний узел не унаследовал список открытых файлов")
	}
}
