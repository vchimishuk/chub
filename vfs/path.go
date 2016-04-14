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
	"strconv"
	"strings"

	"github.com/vchimishuk/chub/cue"
)

const cueExt = "cue"

// Path type is a immutable Virtual File System path representation.
type Path struct {
	root    string
	val     string
	file    string
	dir     bool
	part    bool
	partNum int
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
	pp, n := splitPath(p)
	fp := filePath(root, pp)
	dir, err := isDir(fp)
	if err != nil {
		return nil, err
	}

	return &Path{
		root:    root,
		file:    fp,
		val:     pp,
		dir:     dir,
		part:    n >= 0,
		partNum: n,
	}, nil
}

func (p *Path) Val() string {
	return p.val
}

func (p *Path) File() string {
	return p.file
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
	if p.part {
		return p.val + ":" + strconv.Itoa(p.partNum)
	} else {
		return p.val
	}
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

func (p *Path) ChildPart(name string, number int) (*Path, error) {
	return NewPath(path.Join(p.Val(), name) + ":" + strconv.Itoa(number))
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
	file, err := os.Open(p.File())
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
	file, err := os.Open(p.File())
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
			sheet, err := cue.ParseFile(cp.File())
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
	tracks := make([]Entry, 0)

	for fileNum, file := range sheet.Files {
		f := format(ext(file.Name))
		if f == nil {
			continue
		}

		for _, track := range file.Tracks {
			t, err := cueSheetFileTrack(base, sheet,
				fileNum, track.Number)
			if err == nil {
				tracks = append(tracks, t)
			}
		}
	}

	return tracks, nil
}

func cueSheetFileTrack(base *Path, sheet *cue.Sheet, file int, track int) (*Track, error) {
	if file >= len(sheet.Files) {
		return nil, errors.New("CUE FILE not found")
	}
	f := sheet.Files[file]

	var t *cue.Track
	var ti int
	for i := range f.Tracks {
		if f.Tracks[i].Number == track {
			t, ti = f.Tracks[i], i
			break
		}
	}
	if t == nil {
		return nil, errors.New("CUE TRACK not found")
	}
	pth, err := base.ChildPart(f.Name, track)
	if err != nil {
		return nil, err
	}

	i := startIndex(t.Indexes)
	if i == nil {
		return nil, errors.New("CUE start INDEX not found")
	}

	var start int = i.Time.Seconds()
	var end int

	if ti < len(f.Tracks)-1 {
		ii := startIndex(f.Tracks[ti+1].Indexes)
		if ii == nil {
			return nil, errors.New("CUE start INDEX not found")
		}
		end = ii.Time.Seconds()
	} else {
		format := format(ext(f.Name))
		if format == nil {
			return nil, errors.New("unsupported format")
		}
		end = format.Length(pth.File())
	}

	return &Track{
		Path:   pth,
		Tag:    newTag(sheet, t),
		Length: end - start,
		Part:   true,
		Number: t.Number,
		Start:  start,
		End:    end,
	}, nil
}

func newTrack(p *Path) (*Track, error) {
	f := format(p.Ext())
	if f == nil {
		return nil, errors.New("unsupported format")
	}

	if p.part {
		sheet, err := cueSheetForFile(p)
		// Without CUE information we do not know where track
		// starts and ends, so there is no chance to play it.
		if err != nil {
			return nil, err
		}
		if sheet == nil {
			return nil, errors.New("CUE sheet not found")
		}
		for fn, file := range sheet.Files {
			if file.Name == p.Base() {
				for _, track := range file.Tracks {
					if track.Number == p.partNum {
						base, err := p.Parent()
						if err != nil {
							return nil, err
						}

						return cueSheetFileTrack(base,
							sheet, fn, track.Number)
					}
				}
			}
		}

		return nil, errors.New("track not found")
	} else {
		var err error

		tag, err := f.Tag(p.File())
		if err != nil {
			tag = &Tag{
				Artist: "",
				Album:  "",
				Title:  p.Base(),
				Number: 0,
			}
		}

		return &Track{
			Path:   p,
			Tag:    tag,
			Length: f.Length(p.File()),
		}, nil
	}
}

func isDir(p string) (bool, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

func filePath(root string, p string) string {
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

// startIndex returns Index with number 1, which stands for
// the track start position.
func startIndex(indexes []*cue.Index) *cue.Index {
	for _, i := range indexes {
		if i.Number == 1 {
			return i
		}
	}

	return nil
}

// splitPath splits given VFS path in format FILENAME:TRACK_NUMBER
// to filename and track number parts. Returns -1 as track number
// value if path doesn't represent partial track path.
func splitPath(p string) (string, int) {
	if i := strings.LastIndex(p, ":"); i >= 0 {
		pp := p[0:i]
		n, err := strconv.Atoi(p[i+1:])
		if err != nil || n < 0 {
			return p, -1
		}

		return pp, n
	} else {
		return p, -1
	}
}

// cueSheetForFile returns parsed CUE file for the given audio file.
// Example. Given a FLAC file path it returns a CUE sheet found in the
// same directory which describes the FLAC file.
func cueSheetForFile(p *Path) (*cue.Sheet, error) {
	base := p.Base()
	parent, err := p.Parent()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(parent.File())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if ext(name) == cueExt {
			child, err := parent.Child(name)
			if err != nil {
				continue
			}
			sheet, err := cue.ParseFile(child.File())
			if err != nil {
				return nil, err
			}
			for _, f := range sheet.Files {
				if f.Name == base {
					return sheet, nil
				}
			}
		}
	}

	return nil, nil
}

func newTag(sheet *cue.Sheet, track *cue.Track) *Tag {
	tag := &Tag{
		Album:  sheet.Title,
		Title:  track.Title,
		Number: track.Number,
	}

	if len(track.Performer) > 0 {
		tag.Artist = track.Performer
	} else {
		tag.Artist = sheet.Performer
	}

	return tag
}
