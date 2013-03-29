// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"../config"
	"os"
	"path"
	"strings"
)

// Path type is a Virtual File System path representation.
type Path struct {
	// Root of the VFS.
	root string
	// Directory or File path related to the root.
	filepath string
}

// newPath function returns new Path object.
func newPath(p string) *Path {
	return &Path{root: config.Configurations.VfsRoot(), filepath: p}
}

// copyPath returns copy for the given Path object.
func copyPath(p *Path) *Path {
	return newPath(p.filepath)
}

// Join allows to change path in the same way as shell "cd" command do.
// E.g. if we have path pointed to "/home", and after Join("foo")
// applying path will points to "/home/foo" folder (or file).
func (p *Path) Join(segments ...string) {
	filepath := p.filepath

	for _, segment := range segments {
		if strings.HasPrefix(segment, "/") {
			filepath = segment
		} else {
			filepath = path.Join(filepath, segment)
		}
	}

	p.filepath = path.Clean(filepath)
}

// Stat returns file information for the file pointed of the path object.
func (p *Path) Stat() (fi os.FileInfo, err error) {
	return os.Stat(path.Join(p.root, p.filepath))
}

// Open opens underlying filepath for reading.
func (p *Path) Open() (file *os.File, err error) {
	return os.Open(path.Join(p.root, p.filepath))
}

// String returns string representation of the Path object.
func (p *Path) String() string {
	return p.filepath
}

// Ext returns path's extension part.
func (p *Path) Ext() string {
	return path.Ext(p.filepath)
}

// OsPath returns physical path to the file (in the target OS terms).
func (p *Path) OsPath() string {
	return path.Join(p.root, p.filepath)
}

// ExtMatch returns true if underlying filename matches given extension.
// ext parameter should be in lower case.
func (p *Path) ExtMatch(ext string) bool {
	myExt := strings.ToLower(path.Ext(p.filepath))
	if len(myExt) > 1 {
		myExt = myExt[1:]
	}

	return ext == myExt
}

// Parent returns directory which contains this directory or file.
func (p *Path) Parent() *Path {
	i := strings.LastIndex(p.filepath, "/")

	return newPath(p.filepath[:i])
}
