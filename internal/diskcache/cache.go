package diskcache

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Cache is a simple file-based cache backed by a directory on disk.
type Cache struct {
	dir string
}

// New creates a Cache rooted at dir, creating it if it doesn't exist.
func New(dir string) (*Cache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Cache{dir: dir}, nil
}

// path returns the full path for a named entry.
func (c *Cache) path(name string) string {
	return filepath.Join(c.dir, filepath.Clean("/"+name))
}

// Exists reports whether name is present in the cache.
func (c *Cache) Exists(name string) bool {
	_, err := os.Stat(c.path(name))
	return err == nil
}

// Get returns the contents of name. Returns os.ErrNotExist if not cached.
func (c *Cache) Get(name string) ([]byte, error) {
	data, err := os.ReadFile(c.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, os.ErrNotExist
	}
	return data, err
}

// GetReader returns a ReadCloser for the cached file. Caller must close it.
// Returns os.ErrNotExist if not cached.
func (c *Cache) GetReader(name string) (io.ReadCloser, error) {
	f, err := os.Open(c.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, os.ErrNotExist
	}
	return f, err
}

// Set writes data to the cache under name, creating subdirectories as needed.
func (c *Cache) Set(name string, data []byte) error {
	p := c.path(name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// SetReader writes from r to the cache under name, creating subdirectories as needed.
func (c *Cache) SetReader(name string, r io.Reader) error {
	p := c.path(name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// Delete removes name from the cache. Returns nil if it didn't exist.
func (c *Cache) Delete(name string) error {
	err := os.Remove(c.path(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Path returns the absolute path for name on disk (whether cached or not).
func (c *Cache) Path(name string) string {
	return c.path(name)
}
