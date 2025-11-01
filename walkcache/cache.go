package walkcache

import (
	"io/fs"
	"path/filepath"
	"sync"
)

// Cache is an in-memory implementation of WalkCache.
type Cache struct {
	// Separate maps for files and directories
	files map[string]File
	dirs  map[string][]fs.DirEntry

	// Global locks for map access
	filesMu sync.RWMutex
	dirsMu  sync.RWMutex
}

func (c *Cache) GetFile(path string) (data []byte, info fs.FileInfo, ok bool) {
	path = filepath.Clean(path)
	c.filesMu.RLock()
	defer c.filesMu.RUnlock()
	entry, ok := c.files[path]
	if !ok {
		return nil, FileInfo{}, false
	}
	return entry.data, entry.info, true
}

func (c *Cache) GetDir(path string) (entries []fs.DirEntry, ok bool) {
	path = filepath.Clean(path)
	c.dirsMu.RLock()
	defer c.dirsMu.RUnlock()
	entries, ok = c.dirs[path]
	return entries, ok
}

func (c *Cache) SetFile(path string, data []byte, info fs.FileInfo) {
	path = filepath.Clean(path)
	c.filesMu.Lock()
	defer c.filesMu.Unlock()
	c.files[path] = File{
		data: data,
		info: NewDirEntry(info),
	}
}

func (c *Cache) SetDir(path string, entries []fs.DirEntry) {
	path = filepath.Clean(path)
	c.dirsMu.Lock()
	defer c.dirsMu.Unlock()
	c.dirs[path] = entries
}

// Invalidate removes a path from the cache.
// Removes from both file and directory caches.
func (c *Cache) Invalidate(path string) {
	path = filepath.Clean(path)
	c.filesMu.Lock()
	c.dirsMu.Lock()
	defer func() {
		c.dirsMu.Unlock()
		c.filesMu.Unlock()
	}()
	delete(c.files, path)
	delete(c.dirs, path)
}

// Clear removes all cached data.
func (c *Cache) Clear() {
	c.filesMu.Lock()
	c.dirsMu.Lock()
	defer func() {
		c.filesMu.Unlock()
		c.dirsMu.Unlock()
	}()
	c.files, c.dirs = make(map[string]File), make(map[string][]fs.DirEntry)
}
