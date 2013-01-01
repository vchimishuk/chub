// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"fmt"
	"os"
	"path"
)

// Virtual File System structure.
type Vfs struct {
	// Path to the current working directory.
	wd *Path
}

// New returns newly initialized VFS object.
func New() *Vfs {
	fs := new(Vfs)
	fs.wd = newPath("/")

	return fs
}

// Wd returns current working directory path.
// By analogy with shell (such as Bash) Wd is "pwd" command.
func (fs *Vfs) Wd() *Path {
	return fs.wd
}

// Cd changes current working directory.
// By analogy with shell (such as Bash) Append is "cd dir" command.
func (fs *Vfs) Cd(dir string) error {
	newWd := *fs.wd

	newWd.Join([]string{dir}...)
	fi, err := newWd.Stat()
	if err != nil {
		return fmt.Errorf("Can't change directory. File or directory '%s' does not exist.", newWd.String())
	} else if !fi.IsDir() {
		return fmt.Errorf("Can't change directory. '%s' is not a directory.", newWd.String())
	}

	fs.wd = &newWd

	return nil
}

// List returns current directory content.
// All directory entries (Directory and Track structs) are sorted in the next way:
// alphabetic order sorted directories come first,
// alphabetic sorted files comes after directories list.
func (fs *Vfs) List() (entries []interface{}, err error) {
	// TODO: Directory listing alghorythm here is not optimal and I promice
	//       improve it in the future.

	baseDir := ToOsPath(fs.wd)
	dir, err := os.Open(baseDir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	dirEntries := make(directorySlice, 0, len(names))
	fileEntries := make([]*Track, 0, len(names))

	for _, name := range names {
		filename := path.Join(baseDir, name)
		fi, err := os.Stat(filename)
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			p := copyPath(fs.wd)
			p.Join(name)
			dir := &Directory{Path: p, Name: name}
			dirEntries = append(dirEntries, dir)
		} else if false {
			// TODO:
		} else {
			// Unsupported entry type.
		}
	}

	dirEntries.Sort()

	entries = make([]interface{}, 0, len(dirEntries)+len(fileEntries))
	for _, dir := range dirEntries {
		entries = append(entries, dir)
	}
	for _, file := range fileEntries {
		entries = append(entries, file)
	}

	return entries, nil
}
