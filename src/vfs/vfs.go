// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"../cue"
	"fmt"
	"os"
	"path"
	"sort"
)

const cueExtension = "cue"

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

// NewForPath return VFS object with working directory
// initialized to the given path.
func NewForPath(path *Path) *Vfs {
	fs := New()
	fs.wd = path

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
func (fs *Vfs) Ls() (entries []interface{}, err error) {
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

	dirs, err := fs.readDirs(fs.wd)
	if err != nil {
		return nil, err
	}
	tracks, err := fs.readTracks(fs.wd)
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
func (fs *Vfs) readDirs(dir *Path) (dirs []*Directory, err error) {
	file, err := dir.Open()
	if err != nil {
		return nil, err
	}
	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)

	dirs = make([]*Directory, 0, len(names))

	for _, name := range names {
		fi, err := os.Stat(path.Join(dir.OsPath(), name))
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
func (fs *Vfs) readTracks(dir *Path) (tracks []*Track, err error) {
	file, err := dir.Open()
	if err != nil {
		return nil, err
	}
	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)

	// Process CUE files first and then only audio files which was not
	// found inside CUE ones.

	tracks = make([]*Track, 0, len(names))
	audioFilenames := make([]string, 0, len(names))
	ignoredAudioFilenames := make([]string, 0, len(names))

	for _, name := range names {
		fi, err := os.Stat(path.Join(dir.OsPath(), name))
		if err == nil && !fi.IsDir() {
			file := newPath(dir.String())
			file.Join(name)

			if file.ExtMatch(cueExtension) {
				sheet, err := cue.ParseFile(file.OsPath())
				if err != nil {
					continue
				}
				tracks = append(tracks, fs.parseCueSheet(sheet)...)

				for _, f := range sheet.Files {
					ignoredAudioFilenames = append(
						ignoredAudioFilenames, f.Name)
				}
			} else {
				audioFilenames = append(audioFilenames, name)
			}
		}
	}

	// Now process non-cue audio files.
	sort.Strings(ignoredAudioFilenames)
	l := len(ignoredAudioFilenames)

	for _, filename := range audioFilenames {
		i := sort.SearchStrings(ignoredAudioFilenames, filename)
		if i < l && ignoredAudioFilenames[i] == filename {
			continue
		}

		file := newPath(dir.String())
		file.Join(filename)

		tagReader := NewTagReader(file)
		if tagReader == nil {
			// Unsupported filetype.
			continue
		}
		tag, err := tagReader.Parse(file)
		if err != nil {
			// TODO: Don't ignore files with bad tags,
			//       instead somehow fill tag with
			//       filename information.
			continue
		}

		track := new(Track)
		track.Path = file
		track.Tag = tag
		track.Part = false
		track.Start = 0
		track.End = 0

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// parseCueSheet returns tracks from the given CUE Sheet.
func (fs *Vfs) parseCueSheet(sheet *cue.Sheet) []*Track {
	tracks := make([]*Track, 0, len(sheet.Files))

	for _, file := range sheet.Files {
		p := copyPath(fs.wd)
		p.Join(file.Name)

		for _, track := range file.Tracks {
			tag := new(Tag)
			if len(track.Performer) > 0 {
				tag.Artist = track.Performer
			} else {
				tag.Artist = sheet.Performer
			}
			tag.Album = sheet.Title
			tag.Title = track.Title
			tag.Number = track.Number
			tag.Length = 0

			t := new(Track)
			t.Path = p
			t.Tag = tag
			t.Part = true
			t.Start = 0
			t.End = 0

			tracks = append(tracks, t)
		}

	}

	return tracks
}
