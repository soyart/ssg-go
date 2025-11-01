package walkcache

import (
	"io/fs"
	"time"
)

// WalkCache provides caching for filesystem operations.
// Cache keys are file/directory paths. Paths should be normalized before use.
type WalkCache interface {
	GetFile(path string) (data []byte, info FileInfo, ok bool)
	GetDir(path string) (entries []fs.DirEntry, ok bool)
	SetFile(path string, data []byte, info fs.FileInfo)
	SetDir(path string, entries []fs.DirEntry)

	Invalidate(path string)
	Clear()
}

type File struct {
	data []byte
	info FileInfo
}

// FileInfo implements fs.FileInfo interface.
// It provides cached file information without needing to access the filesystem.
type FileInfo struct {
	modTime  time.Time
	mode     fs.FileMode
	size     int64
	baseName string
}

// Name returns the base name of the file.
func (fi FileInfo) Name() string {
	return fi.baseName
}

// Size returns the length in bytes for regular files.
func (fi FileInfo) Size() int64 {
	return fi.size
}

// Mode returns the file mode bits.
func (fi FileInfo) Mode() fs.FileMode {
	return fi.mode
}

// ModTime returns the modification time.
func (fi FileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir reports whether the file is a directory.
func (fi FileInfo) IsDir() bool {
	return fi.mode.IsDir()
}

// Sys returns the underlying data source (always nil for cached info).
func (fi FileInfo) Sys() any {
	return nil
}

func NewDirEntry(info fs.FileInfo) FileInfo {
	return FileInfo{
		modTime:  info.ModTime(),
		mode:     info.Mode(),
		size:     info.Size(),
		baseName: info.Name(),
	}
}
