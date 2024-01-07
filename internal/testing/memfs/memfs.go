// Package memfs implements the fs.FS interface.
package memfs

import (
	"fmt"
	"io"
	"io/fs"
	"time"
)

type FS struct {
	files map[string][]byte
}

func New(files map[string][]byte) *FS {
	return &FS{files: files}
}

func (f *FS) Open(name string) (fs.File, error) {
	data, ok := f.files[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", fs.ErrNotExist, name)
	}
	return &File{name: name, data: data}, nil
}

type File struct {
	name string
	data []byte
}

func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		name: f.name,
		size: int64(len(f.data)),
	}, nil
}

func (f *File) Read(out []byte) (int, error) {
	n := copy(out, f.data)
	return n, io.EOF
}

func (f *File) Close() error {
	return nil
}

type FileInfo struct {
	name string
	size int64
}

func (fi *FileInfo) Name() string {
	return fi.name
}

func (fi *FileInfo) Size() int64 {
	return fi.size
}

func (fi *FileInfo) Mode() fs.FileMode {
	return fs.FileMode(0644)
}

func (fi *FileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *FileInfo) IsDir() bool {
	return false
}

func (fi *FileInfo) Sys() any {
	return nil
}
