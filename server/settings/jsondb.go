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

var jsonDbLocks = make(map[string]*sync.Mutex)

func NewJsonDB() TorrServerDB {
	settingsDB := &JsonDB{
		Path:              Path,
		filenameDelimiter: ".",
		filenameExtension: ".json",
		fileMode:          fs.FileMode(0o666),
		xPathDelimeter:    "/",
	}
	return settingsDB
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
			if root, err := v.readJsonFileAsMap(filename); err == nil {
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
					return byteData
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
		if root, err := v.readJsonFileAsMap(filename); err == nil {
			delete(root, name)
			if err = v.writeMapAsJsonFile(filename, root); err == nil {
				return
			}
		}
	}
	v.log(fmt.Sprintf("Rem: error removing entry %s->%s", xPath, name), err)
}

func (v *JsonDB) lock(filename string) {
	var mtx sync.Mutex
	if mtx, ok := jsonDbLocks[filename]; !ok {
		mtx = new(sync.Mutex)
		jsonDbLocks[v.Path] = mtx
	}
	mtx.Lock()
}

func (v *JsonDB) unlock(filename string) {
	if mtx, ok := jsonDbLocks[filename]; ok {
		mtx.Unlock()
	}
}

func (v *JsonDB) xPathToFilename(xPath string) (string, error) {
	if pathComponents := strings.Split(xPath, v.xPathDelimeter); len(pathComponents) > 0 {
		return strings.ToLower(strings.Join(pathComponents, v.filenameDelimiter) + v.filenameExtension), nil
	}
	return "", errors.New("xPath has no components")
}

func (v *JsonDB) readJsonFileAsMap(filename string) (map[string]interface{}, error) {
	var err error = nil
	jsonData := map[string]interface{}{}
	path := filepath.Join(v.Path, filename)
	if fileData, err := os.ReadFile(path); err == nil {
		if err = json.Unmarshal(fileData, &jsonData); err != nil {
			v.log(fmt.Sprintf("readJsonFileAsMap(%s) fileData: %s error", filename, fileData), err)
		}
	}
	return jsonData, err
}

func (v *JsonDB) writeMapAsJsonFile(filename string, o map[string]interface{}) error {
	var err error = nil
	path := filepath.Join(v.Path, filename)
	if fileData, err := json.MarshalIndent(o, "", "  "); err == nil {
		if err = os.WriteFile(path, fileData, v.fileMode); err != nil {
			v.log(fmt.Sprintf("writeMapAsJsonFile path: %s, fileMode: %s, fileData: %s error", path, v.fileMode, fileData), err)
		}
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
