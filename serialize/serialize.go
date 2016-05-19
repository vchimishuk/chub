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

package serialize

import (
	"bytes"
	"strconv"

	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/vfs"
)

func Map(m map[string]interface{}) string {
	var b bytes.Buffer
	var l int = len(m)
	var i int = 1

	for k, v := range m {
		b.WriteString(k)
		b.WriteString(": ")

		switch v.(type) {
		case int:
			b.WriteString(strconv.Itoa(v.(int)))
		case string:
			b.WriteString(strconv.Quote(v.(string)))
		case bool:
			b.WriteString(strconv.FormatBool(v.(bool)))
		default:
			panic("Unsupported type")
		}

		if i < l {
			b.WriteString(", ")
		}
		i++
	}

	return b.String()
}

func Entry(entry interface{}) string {
	return Map(entryToMap(entry))
}

func Dir(dir *vfs.Dir) string {
	return Map(dirToMap(dir))
}

func Track(track *vfs.Track) string {
	return Map(trackToMap(track))
}

func PlaylistInfo(pi *player.PlaylistInfo) string {
	return Map(map[string]interface{}{
		"name":     pi.Name(),
		"duration": pi.Duration(),
		"length":   pi.Len(),
	})
}

func entryToMap(e interface{}) map[string]interface{} {
	var m map[string]interface{}

	switch e.(type) {
	case *vfs.Dir:
		m = dirToMap(e.(*vfs.Dir))
		m["type"] = "dir"
	case *vfs.Track:
		m = trackToMap(e.(*vfs.Track))
		m["type"] = "track"
	default:
		panic("unsupported type")
	}

	return m
}

func dirToMap(d *vfs.Dir) map[string]interface{} {
	return map[string]interface{}{
		"path": d.Path.Val(),
		"name": d.Name,
	}
}

func trackToMap(track *vfs.Track) map[string]interface{} {
	m := map[string]interface{}{
		"path":   track.Path.String(),
		"length": track.Length,
	}
	if track.Tag != nil {
		m["artist"] = track.Tag.Artist
		m["album"] = track.Tag.Album
		m["title"] = track.Tag.Title
		m["number"] = track.Tag.Number
	}

	return m
}
