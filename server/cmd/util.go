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

package cmd

import "github.com/vchimishuk/chub/vfs"

func entryToMap(e interface{}) map[string]interface{} {
	switch e.(type) {
	case *vfs.Dir:
		return dirToMap(e.(*vfs.Dir))
	case *vfs.Track:
		return trackToMap(e.(*vfs.Track))
	default:
		panic("unsupported type")
	}
}

func dirToMap(d *vfs.Dir) map[string]interface{} {
	return map[string]interface{}{
		"type": "dir",
		"path": d.Path.Val(),
		"name": d.Name,
	}
}

func trackToMap(t *vfs.Track) map[string]interface{} {
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
	}

	return m
}
