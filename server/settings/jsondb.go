package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"server/log"
)

type JsonDB struct {
	Path              string
	filenameDelimiter string
	filenameExtension string
	fileMode          fs.FileMode
	xPathDelimeter    string
}

var (
	globalJsonDB     TorrServerDB
	jsonDbLocks      = make(map[string]*sync.Mutex)
	jsonDbLocksMutex sync.Mutex
)

func NewJsonDB() TorrServerDB {
	if globalJsonDB != nil {
		return globalJsonDB
	}
	globalJsonDB := &JsonDB{
		Path:              Path,
		filenameDelimiter: ".",
		filenameExtension: ".json",
		fileMode:          fs.FileMode(0o666),
		xPathDelimeter:    "/",
	}
	return globalJsonDB
}

func (v *JsonDB) CloseDB() {
	// Not necessary
}

func (v *JsonDB) Set(xPath, name string, value []byte) {
	var err error = nil
	jsonObj := map[string]interface{}{}
	if err := json.Unmarshal(value, &jsonObj); err == nil {
		if filename, err := v.xPathToFilename(xPath); err == nil {
			v.lock(filename)
			defer v.unlock(filename)
			if root, err := v.readForUpdate(filename); err == nil {
				root[name] = jsonObj
				if err = v.writeMapAsJsonFile(filename, root); err == nil {
					return
				}
			}
		}
	}
	v.log(fmt.Sprintf("Set: error writing entry %s->%s", xPath, name), err)
}

func (v *JsonDB) Get(xPath, name string) []byte {
	var err error = nil
	if filename, err := v.xPathToFilename(xPath); err == nil {
		v.lock(filename)
		defer v.unlock(filename)
		if root, err := v.readJsonFileAsMap(filename); err == nil {
			if jsonData, ok := root[name]; ok {
				if byteData, err := json.Marshal(jsonData); err == nil {
					// Return a copy to be safe
					data := make([]byte, len(byteData))
					copy(data, byteData)
					return data
				}
			} else {
				// We assume this is not 'error' but 'no entry' which is normal
				return nil
			}
		}
	}
	v.log(fmt.Sprintf("Get: error reading entry %s->%s", xPath, name), err)
	return nil
}

func (v *JsonDB) List(xPath string) []string {
	var err error = nil
	if filename, err := v.xPathToFilename(xPath); err == nil {
		v.lock(filename)
		defer v.unlock(filename)
		if root, err := v.readJsonFileAsMap(filename); err == nil {
			nameList := make([]string, 0, len(root))
			for k := range root {
				nameList = append(nameList, k)
			}
			return nameList
		}
	}
	v.log(fmt.Sprintf("List: error reading entries in xPath %s", xPath), err)
	return nil
}

func (v *JsonDB) Rem(xPath, name string) {
	var err error = nil
	if filename, err := v.xPathToFilename(xPath); err == nil {
		v.lock(filename)
		defer v.unlock(filename)
		if root, err := v.readForUpdate(filename); err == nil {
			delete(root, name)
			if err = v.writeMapAsJsonFile(filename, root); err == nil {
				return
			}
		}
	}
	v.log(fmt.Sprintf("Rem: error removing entry %s->%s", xPath, name), err)
}

func (v *JsonDB) Clear(xPath string) {
	filename, err := v.xPathToFilename(xPath)
	if err != nil {
		v.log(fmt.Sprintf("Clear: error converting xPath %s to filename: %v", xPath, err))
		return
	}

	v.lock(filename)
	defer v.unlock(filename)

	path := filepath.Join(v.Path, filename)
	emptyData := []byte("{}")

	if err := os.WriteFile(path, emptyData, v.fileMode); err != nil {
		v.log(fmt.Sprintf("Clear: error writing empty file for xPath %s: %v", xPath, err))
	}
}

func (v *JsonDB) lock(filename string) {
	jsonDbLocksMutex.Lock()
	mtx, ok := jsonDbLocks[filename]
	if !ok {
		mtx = &sync.Mutex{}
		jsonDbLocks[filename] = mtx
	}
	jsonDbLocksMutex.Unlock()
	mtx.Lock()
}

func (v *JsonDB) unlock(filename string) {
	jsonDbLocksMutex.Lock()
	if mtx, ok := jsonDbLocks[filename]; ok {
		mtx.Unlock()
	}
	jsonDbLocksMutex.Unlock()
}

func (v *JsonDB) xPathToFilename(xPath string) (string, error) {
	if pathComponents := strings.Split(xPath, v.xPathDelimeter); len(pathComponents) > 0 {
		return strings.ToLower(strings.Join(pathComponents, v.filenameDelimiter) + v.filenameExtension), nil
	}
	return "", errors.New("xPath has no components")
}

// readJsonFileAsMap returns an empty map only when the file does not exist yet. Read and
// parse errors are returned: the inner err used to shadow the result, so a damaged file
// came back as an empty map, and the next Set wrote that map with a single key over it,
// silently dropping all other settings.
func (v *JsonDB) readJsonFileAsMap(filename string) (map[string]interface{}, error) {
	jsonData := map[string]interface{}{}
	path := filepath.Join(v.Path, filename)
	fileData, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return jsonData, nil
		}
		v.log(fmt.Sprintf("readJsonFileAsMap(%s) read error", filename), err)
		return nil, err
	}
	if err = json.Unmarshal(fileData, &jsonData); err != nil {
		v.log(fmt.Sprintf("readJsonFileAsMap(%s) parse error, %d bytes", filename, len(fileData)), err)
		return nil, fmt.Errorf("%w: %v", errBadJSON, err)
	}
	return jsonData, nil
}

var errBadJSON = errors.New("invalid JSON")

// readForUpdate is readJsonFileAsMap for Set and Rem, which rewrite the file. A file that
// does not parse is moved to <name>.bad, so saving works again and the old content is
// still there to recover from.
func (v *JsonDB) readForUpdate(filename string) (map[string]interface{}, error) {
	root, err := v.readJsonFileAsMap(filename)
	if errors.Is(err, errBadJSON) {
		path := filepath.Join(v.Path, filename)
		if rerr := os.Rename(path, path+".bad"); rerr == nil {
			v.log(fmt.Sprintf("%s is damaged, moved it to %s.bad", filename, filename))
			return map[string]interface{}{}, nil
		}
	}
	return root, err
}

// writeMapAsJsonFile replaces the file atomically: the data goes to a temporary file that
// is synced and then renamed over the old one. Writing in place (os.WriteFile truncates
// first) left a half-written file after a kill or a full disk, and see readJsonFileAsMap
// for what followed.
func (v *JsonDB) writeMapAsJsonFile(filename string, o map[string]interface{}) error {
	path := filepath.Join(v.Path, filename)
	fileData, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		v.log(fmt.Sprintf("writeMapAsJsonFile(%s) marshal error", filename), err)
		return err
	}
	// Callers hold the per-file lock, so a fixed temporary name is safe
	tmp := path + ".tmp"
	err = writeFileSync(tmp, fileData, v.fileMode)
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
		v.log(fmt.Sprintf("writeMapAsJsonFile path: %s, fileMode: %s error", path, v.fileMode), err)
	}
	return err
}

func writeFileSync(name string, data []byte, perm fs.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

func (v *JsonDB) log(s string, params ...interface{}) {
	if len(params) > 0 {
		log.TLogln(fmt.Sprintf("JsonDB: %s: %s", s, fmt.Sprint(params...)))
	} else {
		log.TLogln(fmt.Sprintf("JsonDB: %s", s))
	}
}
