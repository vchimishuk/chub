// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"fmt"
	"os"
	"path"
	"sort"
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
		return fmt.Errorf("File or directory '%s' does not exist.", newWd.String())
	} else if !fi.IsDir() {
		return fmt.Errorf("'%s' is not a directory.", newWd.String())
	}

	fs.wd = &newWd

	return nil
}

// List returns current directory content.
// All directory entries (Directory and Track structs) are sorted in the next way:
// alphabetic order sorted directories come first,
// alphabetic sorted files comes after directories list.
func (fs *Vfs) List() (entries []interface{}, err error) {
	// TODO: Directory listing alghorythm here is not optimal
	//       and I promice improve it in the future.

	// Listing algorythm works in the next simple way. Current folder
	// scans three times:
	// 1. On the first iteration only directories are selected and
	//    collected into directories list.
	// 2. With second one only cue files are selected and parsed.
	//    Result stored in traks list.
	// 3. With last iteration all other supported audio files are selected
	//    and appended to the traks lis. But audio files which were added
	//    with step number are ignored, becuase they were processed from
	//    cue files parsing step and we don't need them appear second
	//    time in list.
	// Result entries list is formed from selected directories list
	// and selected tracks list appended.

	wd, err := fs.wd.Open()
	if err != nil {
		return nil, err
	}
	defer wd.Close()

	dirs, err := fs.readDirs(wd)
	if err != nil {
		return nil, err
	}
	tracks, err := fs.readTracks(wd)
	if err != nil {
		return nil, err
	}

	entries = make([]interface{}, 0, len(dirs)+len(tracks))
	for _, dir := range dirs {
		entries = append(entries, dir)
	}
	for _, track := range tracks {
		entries = append(entries, track)
	}

	return entries, nil
}

// readDirs returns sorted list of all directories for the given file object.
// Parameter is garanteed to be a folder.
func (fs *Vfs) readDirs(dir *os.File) (dirs []*Directory, err error) {
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)

	dirs = make([]*Directory, 0, len(names))

	for _, name := range names {
		fi, err := os.Stat(path.Join(dir.Name(), name))
		if err == nil && fi.IsDir() {
			p := copyPath(fs.wd)
			p.Join(name)
			d := &Directory{Path: p, Name: name}
			dirs = append(dirs, d)
		}
	}

	return dirs, nil
}

// readTracks returns sorted tracks list in the given directory.
// Parameter is garanteed to be a folder.
func (fs *Vfs) readTracks(dir *os.File) (tracks []*Track, err error) {
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)

	tracks = make([]*Track, 0, len(names))

	for _, name := range names {
		fi, err := os.Stat(path.Join(dir.Name(), name))
		if err == nil && !fi.IsDir() {
			// TODO:
		}
	}

	return tracks, nil
}
