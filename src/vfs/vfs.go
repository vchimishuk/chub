// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"fmt"
)

// Basic filesystem entry structure.
type entry struct {
	// Path to the file.
	Path *Path
}

// Directory is a filesystem entry structure which representing
// directories.
type Drirectory struct {
	// Basic entry specific things.
	entry
	// Directory name.
	Name string
}

// Track is a filesystem entry structure representing track.
// In VFS terms track can represents whole file (usually MP3 file
// in some particular folder equals one track from an album), or piece
// of some physical file (e. g. FLAC file can consists from many tracks).
type Track struct {
	// Basic entry specific things.
	entry
	// Track media information.
	Tag int // XXX: ??? //*audio.Tag
	// If Part is true it means this track represents piece of the 
	// actual file, not the whole file. E. g. we have album FLAC file
	// and this track points only to one song in this FLAC file (which is
	// album itself).
	Part bool
	// Track beginning in the physical file.
	Start int
	// Track end position in the physical file.
	End int
}

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
