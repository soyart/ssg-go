package walkcache

import (
	"io/fs"
)

// Filesystem provides proxy for functions from os package
type Filesystem interface {
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	Stat(path string) (fs.FileInfo, error)
}

func New() Filesystem {
	return &proxy{
		Cache{
			files: make(map[string]File),
			dirs:  make(map[string][]fs.DirEntry),
		},
	}
}

// WalkCache provides caching for filesystem operations.
// Cache keys are file/directory paths. Paths should be normalized before use.
type WalkCache interface {
	GetFile(path string) (data []byte, info fs.FileInfo, ok bool)
	GetDir(path string) (entries []fs.DirEntry, ok bool)
	SetFile(path string, data []byte, info fs.FileInfo)
	SetDir(path string, entries []fs.DirEntry)

	Invalidate(path string)
	Clear()
}

type File struct {
	data []byte
	info fs.FileInfo
}
