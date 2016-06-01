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
	tracks   *Tracks
}

func NewPlaylist(name string) *Playlist {
	return &Playlist{
		name:   name,
		tracks: &Tracks{},
	}
}

func (pl *Playlist) Name() string {
	return pl.name
}

func (pl *Playlist) SetName(name string) {
	pl.name = name
}

func (pl *Playlist) Duration() int {
	return pl.duration
}

func (pl *Playlist) Tracks() *Tracks {
	return pl.tracks
}

func (pl *Playlist) Len() int {
	return pl.tracks.Len()
}

func (pl *Playlist) Clear() {
	pl.tracks = &Tracks{}
	pl.duration = 0
}

func (pl *Playlist) Find(track *vfs.Track) int {
	for i := 0; i < pl.tracks.Len(); i++ {
		if &pl.tracks.Get(i).Path == &track.Path {
			return i
		}
	}

	return -1
}

func (pl *Playlist) Append(tracks ...*vfs.Track) {
	pl.tracks = pl.tracks.Append(tracks...)
	for _, t := range tracks {
		pl.duration += t.Length
	}
}

func (pl *Playlist) clone() *Playlist {
	return &Playlist{
		name:     pl.name,
		duration: pl.duration,
		tracks:   pl.tracks,
	}
}
