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
		dir:  dir,
		hash: make(map[string]*fileChange),
	}
}

type FileHash struct {
	dir  string
	hash map[string]*fileChange
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
		h.hash[key] = &fileChange{
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

	m := md5.New()

	reader, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer reader.Close()

	_, err = io.Copy(m, reader)
	if err != nil {
		return false, err
	}

	sum := m.Sum(nil)
	value := hex.EncodeToString(sum[:])

	_, name := filepath.Split(path)
	if val, ok := h.hash[name]; !ok || val.hash != value {
		h.hash[name] = &fileChange{
			hash:   value,
			change: true,
		}
		return true, nil
	}
	if h.hash[name].change {
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
			h.hash[name].change = true
			return true, nil
		}
	}
	return false, nil
}

func (h *FileHash) Save() error {
	var temp = make(map[string]string)
	for key, val := range h.hash {
		temp[key] = val.hash
	}

	marshal, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(h.dir, ".build.json"), marshal, 0644)
}
