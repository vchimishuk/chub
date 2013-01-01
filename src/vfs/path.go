// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"os"
	"path"
	"strings"
)

const (
	ROOT = "/home" // TODO: This value should be taken form configurations.
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
	return &Path{root: ROOT, filepath: p}
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

// String returns string representation of the Path object.
func (p *Path) String() string {
	return p.filepath
}

// ToOsPath returns physical path to the file (in the target OS terms).
func ToOsPath(p *Path) string {
	return path.Join(p.root, p.filepath)
}
