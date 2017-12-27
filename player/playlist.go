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

package player

import "github.com/vchimishuk/chub/vfs"

type Playlist struct {
	name     string
	duration int
	tracks   []*vfs.Track
}

func NewPlaylist(name string) *Playlist {
	return &Playlist{name: name}
}

func (pl *Playlist) Name() string {
	return pl.name
}

func (pl *Playlist) SetName(name string) *Playlist {
	return &Playlist{name: name, duration: pl.duration, tracks: pl.tracks}
}

func (pl *Playlist) Duration() int {
	return pl.duration
}

func (pl *Playlist) Len() int {
	return len(pl.tracks)
}

func (pl *Playlist) Clear() *Playlist {
	return &Playlist{name: pl.name, duration: 0, tracks: nil}
}

func (pl *Playlist) Get(i int) *vfs.Track {
	return pl.tracks[i]
}

func (pl *Playlist) Append(tracks ...*vfs.Track) *Playlist {
	t := make([]*vfs.Track, len(pl.tracks)+len(tracks))
	copy(t, pl.tracks)
	copy(t[len(pl.tracks):], tracks)

	d := pl.duration
	for _, t := range tracks {
		d += t.Length
	}

	return &Playlist{name: pl.name, duration: d, tracks: t}
}
