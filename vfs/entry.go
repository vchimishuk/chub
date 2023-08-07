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
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

package vfs

import "github.com/vchimishuk/chub/serialize"

type Entry interface {
	serialize.Serializable
	IsDir() bool
	Dir() *Dir
	Track() *Track
}

// Dir is a filesystem entry structure which representing
// directories.
type Dir struct {
	// Path to the file.
	Path *Path
	// Directory name.
	Name string
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) Dir() *Dir {
	return d
}

func (d *Dir) Track() *Track {
	panic(nil)
}

func (d *Dir) Serialize() string {
	return serialize.Map(map[string]interface{}{
		"type": "dir",
		"path": d.Path.Val(),
		"name": d.Name,
	})
}

// Track's tag data.
type Tag struct {
	// Artist name.
	Artist string
	// Album name.
	Album string
	// Album release year.
	Year int
	// Track's title.
	Title string
	// Track number.
	Number int
}

// Track is a filesystem entry structure representing track.
// In VFS terms track can represents whole file (usually MP3 file
// in some particular folder equals one track from an album), or piece
// of some physical file (e. g. FLAC file can consists from many tracks).
type Track struct {
	// Path to the file.
	Path *Path
	// Track media information.
	Tag *Tag
	// Track length in seconds.
	// TODO: Rename to Duration to be consistent with Playlist.
	Length int
	// If Part is true it means this track represents piece of the
	// actual file, not the whole file. E. g. we have album FLAC file
	// and this track points only to one song in this FLAC file (which is
	// album itself).
	Part bool
	// Track number as in CUE file.
	Number int
	// Track beginning in the physical file.
	Start int
	// Track end position in the physical file.
	End int
}

func (t *Track) IsDir() bool {
	return false
}

func (t *Track) Dir() *Dir {
	panic(nil)
}

func (t *Track) Track() *Track {
	return t
}

func (t *Track) Serialize() string {
	m := map[string]interface{}{
		"type":   "track",
		"path":   t.Path.String(),
		"length": t.Length,
	}
	if t.Tag != nil {
		m["artist"] = t.Tag.Artist
		m["album"] = t.Tag.Album
		m["title"] = t.Tag.Title
		m["number"] = t.Tag.Number
		m["year"] = t.Tag.Year
	}

	return serialize.Map(m)
}
