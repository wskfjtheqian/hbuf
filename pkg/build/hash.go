package build

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func NewFileHash(dir string) *FileHash {
	return &FileHash{
		dir:  dir,
		hash: make(map[string]string),
	}
}

type FileHash struct {
	dir  string
	hash map[string]string
}

func (h *FileHash) Read() error {
	file, err := os.ReadFile(filepath.Join(h.dir, ".build.json"))
	if err != nil {
		return nil
	}
	return json.Unmarshal(file, &h.hash)
}

func (h *FileHash) Check(path string) (bool, error) {
	m := md5.New()

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = io.Copy(m, file)
	if err != nil {
		return false, err
	}

	sum := m.Sum(nil)
	value := hex.EncodeToString(sum[:])

	_, name := filepath.Split(path)
	if h.hash[name] == value {
		return true, nil
	}

	h.hash[name] = value
	return false, nil
}

func (h *FileHash) Save() error {
	marshal, err := json.MarshalIndent(h.hash, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(h.dir, ".build.json"), marshal, 0644)
}
