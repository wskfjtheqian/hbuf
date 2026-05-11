package build

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hbuf/pkg/ast"
	"io"
	"os"
	"path/filepath"
)

type fileChange struct {
	hash   string
	change bool
}

func NewFileHash(dir string) *FileHash {
	return &FileHash{
		dir:       dir,
		changeMap: make(map[string]*fileChange),
		hashMap:   make(map[string]string),
	}
}

type FileHash struct {
	dir       string
	changeMap map[string]*fileChange
	hashMap   map[string]string
}

func (h *FileHash) Read() error {
	file, err := os.ReadFile(filepath.Join(h.dir, ".build.json"))
	if err != nil {
		return nil
	}
	var temp map[string]string
	err = json.Unmarshal(file, &temp)
	if err != nil {
		return err
	}
	for key, val := range temp {
		h.changeMap[key] = &fileChange{
			hash:   val,
			change: false,
		}
	}
	return nil
}

func (h *FileHash) CheckChange(path string, files map[string]*ast.File, parents ...string) (bool, error) {

	for _, item := range parents {
		if item == path {
			return false, nil
		}
	}
	value := h.hashMap[path]
	_, name := filepath.Split(path)
	if val, ok := h.changeMap[name]; !ok || val.hash != value {
		h.changeMap[name] = &fileChange{
			hash:   value,
			change: true,
		}
		return true, nil
	}
	if h.changeMap[name].change {
		return true, nil
	}

	file, ok := files[path]
	if !ok {
		return false, nil
	}

	for _, item := range file.Imports {
		change, err := h.CheckChange(item.Path.Value, files, append(parents, path)...)
		if err != nil {
			return false, err
		}
		if change {
			h.changeMap[name].change = true
			return true, nil
		}
	}
	return false, nil
}

func (h *FileHash) Save() error {
	var temp = make(map[string]string)
	for key, val := range h.changeMap {
		temp[key] = val.hash
	}

	marshal, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(h.dir, ".build.json"), marshal, 0644)
}

func (h *FileHash) Init(files map[string]*ast.File) error {
	for key, _ := range files {
		err := func(path string) error {
			m := md5.New()

			reader, err := os.Open(path)
			if err != nil {
				return err
			}
			defer reader.Close()

			_, err = io.Copy(m, reader)
			if err != nil {
				return err
			}

			sum := m.Sum(nil)
			h.hashMap[path] = hex.EncodeToString(sum[:])
			return nil
		}(key)
		if err != nil {
			return err
		}
	}
	return nil
}
