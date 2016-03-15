// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of Chub.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub.  If not, see <http://www.gnu.org/licenses/>.

// vfs package provides convenient interface over filesystem and audio files.
// It allows navigation over multiple-track files as tracks are separate files.
package vfs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vchimishuk/chub/cue"
)

const cueExt = "cue"

// Path type is a immutable Virtual File System path representation.
type Path struct {
	root string
	val  string
	full string
	dir  bool
}

var root string = "/"

func SetRoot(dir string) error {
	d, err := isDir(dir)
	if err != nil {
		return err
	}
	if !d {
		return errors.New("not a directory")
	}
	root = dir

	return nil
}

func NewPath(p string) (*Path, error) {
	fp := fullPath(root, p)
	dir, err := isDir(fp)
	if err != nil {
		return nil, err
	}

	return &Path{root: root, full: fp, val: p, dir: dir}, nil
}

func (p *Path) Val() string {
	return p.val
}

func (p *Path) Full() string {
	return p.full
}

func (p *Path) Parent() (*Path, error) {
	return NewPath(path.Dir(p.val))
}

func (p *Path) Base() string {
	return path.Base(p.val)
}

func (p *Path) IsDir() bool {
	return p.dir
}

func (p *Path) String() string {
	return p.val
}

func (p *Path) Dir() (*Dir, error) {
	if !p.IsDir() {
		return nil, fmt.Errorf("'%s 'is not directory", p)
	}

	return &Dir{Path: p, Name: p.Base()}, nil
}

func (p *Path) Track() (*Track, error) {
	if p.IsDir() {
		return nil, fmt.Errorf("'%s' is not track", p)
	}

	return newTrack(p)
}

func (p *Path) Child(name string) (*Path, error) {
	return NewPath(path.Join(p.Val(), name))
}

// List returns current directory contents sorted.
// All directory entries (Directory and Track structs) are sorted in the next way:
// alphabetic order sorted directories come first,
// alphabetic sorted files comes after directories list.
func (p *Path) List() ([]Entry, error) {
	// TODO: Directory listing alghorythm here is not optimal
	//       and I promise to improve it in the future.

	// Listing algorythm works in the next simple way. Current folder
	// scans three times:
	// 1. On the first iteration only directories are selected and
	//    collected into directories list.
	// 2. With second one only cue files are selected and parsed.
	//    Result stored in tracks list.
	// 3. With last iteration all other supported audio files are selected
	//    and appended to the tracks lis. But audio files which were added
	//    with step number two are ignored, becuase they were processed from
	//    cue files parsing step and we don't need them appear second
	//    time in list.
	// Result entries list is formed from selected directories list
	// and selected tracks list appended.

	if !p.IsDir() {
		return nil, fmt.Errorf("'%s' is not directory", p)
	}

	dirs, err := readDirs(p)
	if err != nil {
		return nil, err
	}
	tracks, err := readTracks(p)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(dirs)+len(tracks))
	entries = append(entries, dirs...)
	entries = append(entries, tracks...)

	return entries, nil
}

func (p *Path) Ext() string {
	return ext(p.val)
}

// readDirs returns sorted list of all directories for the given path.
// Parameter is garanteed to be a folder.
func readDirs(p *Path) ([]Entry, error) {
	file, err := os.Open(p.Full())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	dirs := make([]Entry, 0, len(names))

	for _, name := range names {
		cp, err := p.Child(name)
		if err != nil {
			return nil, err
		}
		if cp.IsDir() {
			dirs = append(dirs, &Dir{Path: cp, Name: name})
		}
	}

	return dirs, nil
}

// readTracks returns sorted tracks list in the given directory.
// Parameter is garanteed to be a folder.
func readTracks(p *Path) ([]Entry, error) {
	file, err := os.Open(p.Full())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	sort.Strings(names)

	// Process CUE files first and then only audio files which was not
	// found inside CUE ones.
	tracks := make([]Entry, 0, len(names))
	// Standalone audio files.
	standaloneFiles := make([]string, 0, len(names))
	// Files used in CUEs. As far as tracks for them were generated from
	// corresponding CUE file they must not be processed as standalone.
	cueFiles := make([]string, 0, len(names))

	for _, name := range names {
		cp, err := p.Child(name)
		if err != nil {
			return nil, err
		}
		if cp.IsDir() {
			continue
		}
		if cp.Ext() == cueExt {
			sheet, err := cue.ParseFile(cp.Full())
			if err != nil {
				continue
			}
			cueTracks, err := cueSheetTracks(p, sheet)
			if err != nil {
				continue
			}
			tracks = append(tracks, cueTracks...)

			for _, f := range sheet.Files {
				cueFiles = append(
					cueFiles, f.Name)
			}
		} else {
			standaloneFiles = append(standaloneFiles, name)
		}
	}

	// Now process non-cue audio files.
	sort.Strings(cueFiles)

	for _, name := range standaloneFiles {
		if stringsPresent(cueFiles, name) {
			continue
		}
		pp, err := p.Child(name)
		if err != nil {
			// TODO: Log ignored file.
		} else {
			ext := strings.ToLower(pp.Ext())
			if format(ext) != nil {
				track, err := pp.Track()
				if err != nil {
					panic("not track")
				}
				tracks = append(tracks, track)
			}
		}
	}

	return tracks, nil
}

func cueSheetTracks(base *Path, sheet *cue.Sheet) ([]Entry, error) {
	tracks := make([]Entry, 0, len(sheet.Files))

	for _, file := range sheet.Files {
		ext := ext(file.Name)
		if format(ext) == nil {
			continue
		}

		p, err := base.Child(file.Name)
		if err != nil {
			return nil, err
		}

		for _, track := range file.Tracks {
			tag := &Tag{}
			if len(track.Performer) > 0 {
				tag.Artist = track.Performer
			} else {
				tag.Artist = sheet.Performer
			}
			tag.Album = sheet.Title
			tag.Title = track.Title
			tag.Number = track.Number

			t := &Track{}
			t.Path = p
			t.Tag = tag
			t.Length = 0
			t.Part = true
			t.Start = 0 // TODO:
			t.End = 0   // TODO:

			tracks = append(tracks, t)
		}

	}

	return tracks, nil
}

func newTrack(p *Path) (*Track, error) {
	f := format(p.Ext())
	if f == nil {
		return nil, errors.New("unsupported format")
	}

	tag, err := f.Tag(p.Full())
	if err != nil {
		// TODO: Fill tag based on file name?
		tag = &Tag{
			Artist: "",
			Album:  "",
			Title:  "",
			Number: 0,
		}
	}

	return &Track{Path: p, Tag: tag, Length: f.Length(p.Full())}, nil
}

func isDir(p string) (bool, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

func fullPath(root string, p string) string {
	fp := path.Join(root, filepath.Clean(p))

	// Be sure that we have not escaped from the root.
	if !strings.HasPrefix(fp, root) {
		fp = root
	}

	return fp
}

func stringsPresent(haystack []string, needle string) bool {
	i := sort.SearchStrings(haystack, needle)

	return i < len(haystack) && haystack[i] == needle
}

func ext(p string) string {
	ext := strings.ToLower(path.Ext(p))
	if len(ext) > 0 {
		ext = ext[1:]
	}

	return ext
}
