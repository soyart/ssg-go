package walkcache

import (
	"io/fs"
	"os"
)

// OSProxy provides proxy for functions from os package
type OSProxy interface {
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	Stat(path string) (fs.FileInfo, error)
}

type proxy struct {
	Cache
}

func NewOS() OSProxy {
	return &proxy{
		Cache{
			files: make(map[string]File),
			dirs:  make(map[string][]fs.DirEntry),
		},
	}
}

func (w *proxy) ReadFile(path string) ([]byte, error) {
	data, info, ok := w.GetFile(path)
	if ok {
		return data, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	w.SetFile(path, data, info)
	return data, nil
}

func (w *proxy) ReadDir(path string) ([]fs.DirEntry, error) {
	data, ok := w.GetDir(path)
	if ok {
		return data, nil
	}
	data, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	w.SetDir(path, data)
	return data, nil
}

func (w *proxy) Stat(path string) (fs.FileInfo, error) {
	data, info, ok := w.GetFile(path)
	if ok {
		return info, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	w.SetFile(path, data, info)
	return info, nil
}
