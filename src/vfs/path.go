// vfs package provides convenient interface over filesystem and audio files.
// It allows navigation over multiple-track files as tracks are separate files.
package vfs

import (
	"../config"
	"../cue"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

const cueExtension = "cue"

// fileInfo saves a copy of os.FileInfo.
type fileInfo struct {
	isDirectory bool
	name        string
}

// Path type is a Virtual File System path representation.
// Path object is immutable.
type Path struct {
	// Root of the VFS.
	root string
	// Directory or File path related to the root.
	filepath string
	// Cache os.FileInfo object to not be annoying with syscall to much.
	fi *fileInfo
}

// NewPath function returns new Path object.
func NewPath(p string) *Path {
	return &Path{root: config.Configurations.VfsRoot(),
		filepath: p,
		fi:       nil}
}

// Join joins given path elements to the Path and returns brand new Path object.
func (p *Path) Join(segments ...string) *Path {
	filepath := p.filepath

	for _, segment := range segments {
		if strings.HasPrefix(segment, "/") {
			filepath = segment
		} else {
			filepath = path.Join(filepath, segment)
		}
	}

	return NewPath(path.Clean(filepath))
}

// Stat returns file information for the file pointed of the path object.
func (p *Path) stat() (fi *fileInfo, err error) {
	if p.fi == nil {
		fi, err := os.Stat(path.Join(p.root, p.filepath))
		if err != nil {
			return nil, err
		}

		p.fi = new(fileInfo)
		p.fi.isDirectory = fi.IsDir()
		p.fi.name = fi.Name()
	}

	return p.fi, err
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

	return NewPath(p.filepath[:i])
}

// IsDirectory return true if pointed object is directory.
func (p *Path) IsDirectory() (dir bool, err error) {
	fi, err := p.stat()
	if err != nil {
		return false, err
	}

	return fi.isDirectory, nil
}

// Directory return Directory object represented by the Path.
func (p *Path) Directory() (dir *Directory, err error) {
	fi, err := p.stat()
	if err != nil {
		return nil, err
	}
	if !fi.isDirectory {
		panic("This path doesn't represents a directory.")
	}

	return &Directory{Path: p, Name: fi.name}, nil
}

// Track return Track represented by the Path object.
func (p *Path) Track() (track *Track, err error) {
	fi, err := p.stat()
	if err != nil {
		return nil, err
	}
	if fi.isDirectory {
		panic("This path doesn't represents a track.")
	}

	return newTrack(p)
}

// List returns current directory content.
// All directory entries (Directory and Track structs) are sorted in the next way:
// alphabetic order sorted directories come first,
// alphabetic sorted files comes after directories list.
func (p *Path) List() (entries []interface{}, err error) {
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

	isDir, err := p.IsDirectory()

	if err != nil {
		return nil, err
	}
	if !isDir {
		panic("Path is not a directory, so can't be listed.")
	}

	dirs, err := readDirs(p)
	if err != nil {
		return nil, err
	}
	tracks, err := readTracks(p)
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
func readDirs(dir *Path) (dirs []*Directory, err error) {
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
			p := dir.Join(name)
			d := &Directory{Path: p, Name: name}
			dirs = append(dirs, d)
		}
	}

	return dirs, nil
}

// readTracks returns sorted tracks list in the given directory.
// Parameter is garanteed to be a folder.
func readTracks(dir *Path) (tracks []*Track, err error) {
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
			file := NewPath(dir.String())
			file.Join(name)

			if file.ExtMatch(cueExtension) {
				sheet, err := cue.ParseFile(file.OsPath())
				if err != nil {
					continue
				}
				tracks = append(tracks, parseCueSheet(dir, sheet)...)

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

		file := dir.Join(filename)
		track, err := newTrack(file)
		if err != nil {
			// TODO: Log it.
		} else {
			tracks = append(tracks, track)
		}
	}

	return tracks, nil
}

// newTrack creates brand new Track represented by path.
func newTrack(path *Path) (track *Track, err error) {
	tagReader := NewTagReader(path)
	if tagReader == nil {
		// Unsupported filetype.
		return nil, fmt.Errorf("Unsupported track %s.", path.String())
	}
	tag, err := tagReader.Parse(path)
	if err != nil {
		// TODO: Don't ignore files with bad tags,
		//       instead somehow fill tag with
		//       filename information.
		return nil, fmt.Errorf("Bad tagged track %s.", path.String())
	}

	track = new(Track)
	track.Path = path
	track.Tag = tag
	track.Part = false
	track.Start = 0
	track.End = 0

	return track, nil
}

// parseCueSheet returns tracks from the given CUE Sheet.
func parseCueSheet(dir *Path, sheet *cue.Sheet) []*Track {
	tracks := make([]*Track, 0, len(sheet.Files))

	for _, file := range sheet.Files {
		p := dir.Join(file.Name)

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
