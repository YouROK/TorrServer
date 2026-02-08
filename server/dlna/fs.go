package dlna

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/anacrolix/dms/dlna"
	"github.com/anacrolix/dms/upnpav"

	"server/log"
	mt "server/mimetype"
	"server/settings"
)

func localDLNARootDir() (string, error) {
	if settings.BTsets == nil {
		return "", fmt.Errorf("settings not initialized")
	}

	if settings.BTsets.DLNALocalRoot != "" {
		return settings.BTsets.DLNALocalRoot, nil
	}

	// Fallback to executable directory ("where server is installed")
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// dlnaPath is "/FS" or "/FS/sub/dir"
func fsRelFromDLNAPath(dlnaPath string) (string, error) {
	if dlnaPath == "/FS" {
		return "", nil
	}
	if !strings.HasPrefix(dlnaPath, "/FS/") {
		return "", fmt.Errorf("not a FS path: %s", dlnaPath)
	}
	rel := strings.TrimPrefix(dlnaPath, "/FS/")
	rel = filepath.FromSlash(rel)
	rel = filepath.Clean(rel)
	if rel == "." {
		rel = ""
	}
	return rel, nil
}

func secureJoin(root, rel string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	full := filepath.Join(rootAbs, rel)
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}

	if fullAbs != rootAbs {
		prefix := rootAbs + string(os.PathSeparator)
		if !strings.HasPrefix(fullAbs, prefix) {
			return "", fmt.Errorf("path escapes root")
		}
	}

	return fullAbs, nil
}

func browseFS(dlnaPath, host string) (ret []interface{}, err error) {
	if settings.BTsets == nil || !settings.BTsets.EnableDLNALocal {
		return nil, nil
	}

	root, err := localDLNARootDir()
	if err != nil {
		return nil, err
	}

	rel, err := fsRelFromDLNAPath(dlnaPath)
	if err != nil {
		return nil, err
	}

	full, err := secureJoin(root, rel)
	if err != nil {
		return nil, err
	}

	st, err := os.Stat(full)
	if err != nil {
		return nil, err
	}

	// If a file is browsed directly, return its item.
	if !st.IsDir() {
		item, ok := makeItemFromLocalFile(dlnaPath, host, full, st)
		if ok {
			ret = append(ret, item)
		}
		return
	}

	entries, err := os.ReadDir(full)
	if err != nil {
		return nil, err
	}

	// Deterministic order: dirs first, then files; both alphabetical.
	type wrap struct {
		e     os.DirEntry
		name  string
		isDir bool
	}
	list := make([]wrap, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		list = append(list, wrap{e: e, name: name, isDir: e.IsDir()})
	}

	sort.SliceStable(list, func(i, j int) bool {
		if list[i].isDir != list[j].isDir {
			return list[i].isDir
		}
		return strings.ToLower(list[i].name) < strings.ToLower(list[j].name)
	})

	currentID := url.PathEscape(dlnaPath)

	for _, w := range list {
		name := w.name

		childDlnaPath := dlnaPath
		if childDlnaPath == "/FS" {
			childDlnaPath = "/FS/" + name
		} else {
			childDlnaPath = dlnaPath + "/" + name
		}

		if w.isDir {
			obj := upnpav.Object{
				ID:         url.PathEscape(childDlnaPath),
				ParentID:   currentID,
				Restricted: 1,
				Title:      name,
				Class:      "object.container.storageFolder",
				Date:       upnpav.Timestamp{Time: time.Now()},
			}
			cnt := upnpav.Container{Object: obj, ChildCount: 0}
			ret = append(ret, cnt)
			continue
		}

		info, err := w.e.Info()
		if err != nil {
			continue
		}

		fullChild, err := secureJoin(root, filepath.Join(rel, filepath.FromSlash(name)))
		if err != nil {
			continue
		}

		item, ok := makeItemFromLocalFile(childDlnaPath, host, fullChild, info)
		if !ok {
			continue
		}
		item.ParentID = currentID
		ret = append(ret, item)
	}

	return
}

func getFSMetadata(dlnaPath, host string) (ret interface{}, err error) {
	if settings.BTsets == nil || !settings.BTsets.EnableDLNALocal {
		return nil, fmt.Errorf("local dlna disabled")
	}

	root, err := localDLNARootDir()
	if err != nil {
		return nil, err
	}

	rel, err := fsRelFromDLNAPath(dlnaPath)
	if err != nil {
		return nil, err
	}

	full, err := secureJoin(root, rel)
	if err != nil {
		return nil, err
	}

	st, err := os.Stat(full)
	if err != nil {
		return nil, err
	}

	if st.IsDir() {
		title := "Local files"
		if dlnaPath != "/FS" {
			title = filepath.Base(dlnaPath)
		}

		obj := upnpav.Object{
			ID:         url.PathEscape(dlnaPath),
			ParentID:   url.PathEscape(filepath.Dir(dlnaPath)),
			Restricted: 1,
			Searchable: 1,
			Title:      title,
			Class:      "object.container.storageFolder",
			Date:       upnpav.Timestamp{Time: time.Now()},
		}
		meta := upnpav.Container{Object: obj, ChildCount: 0}
		return meta, nil
	}

	item, ok := makeItemFromLocalFile(dlnaPath, host, full, st)
	if !ok {
		return nil, fmt.Errorf("unsupported file type")
	}
	return item, nil
}

func makeItemFromLocalFile(dlnaPath, host, fullPath string, st os.FileInfo) (item upnpav.Item, ok bool) {
	mime, err := mt.MimeTypeByPath(fullPath)
	if err != nil {
		if settings.BTsets != nil && settings.BTsets.EnableDebug {
			log.TLogln("Can't detect mime type", err)
		}
		return upnpav.Item{}, false
	}

	// Same behavior as torrents: only media
	if !mime.IsMedia() {
		return upnpav.Item{}, false
	}

	obj := upnpav.Object{
		ID:         url.PathEscape(dlnaPath),
		ParentID:   url.PathEscape(filepath.Dir(dlnaPath)),
		Restricted: 1,
		Title:      filepath.Base(fullPath),
		Class:      "object.item." + mime.Type() + "Item",
		Date:       upnpav.Timestamp{Time: time.Now()},
	}

	item = upnpav.Item{
		Object: obj,
		Res:    make([]upnpav.Resource, 0, 1),
	}

	rel, err := fsRelFromDLNAPath(dlnaPath)
	if err != nil {
		return upnpav.Item{}, false
	}

	// IMPORTANT: path-based endpoint; more compatible than query params for some TVs.
	resourceURL := getLink(host, "dlna/fs/"+url.PathEscape(filepath.ToSlash(rel)))

	item.Res = append(item.Res, upnpav.Resource{
		URL: resourceURL,
		ProtocolInfo: fmt.Sprintf("http-get:*:%s:%s", mime, dlna.ContentFeatures{
			SupportRange:    true,
			SupportTimeSeek: true,
		}.String()),
		Size: uint64(st.Size()),
	})

	return item, true
}
