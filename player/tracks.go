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

import (
	"github.com/vchimishuk/chub/vfs"
)

type Tracks struct {
	s []*vfs.Track
}

func newTracks(tracks []*vfs.Track) *Tracks {
	return &Tracks{s: tracks}
}

func (t *Tracks) Len() int {
	return len(t.s)
}

func (t *Tracks) Get(i int) *vfs.Track {
	return t.s[i]
}

func (t *Tracks) Set(i int, track *vfs.Track) *Tracks {
	ss := make([]*vfs.Track, len(t.s))
	copy(ss, t.s)
	ss[i] = track

	return &Tracks{s: ss}
}

func (t *Tracks) Insert(i int, tracks ...*vfs.Track) *Tracks {
	if len(t.s) == 0 || i == len(t.s) {
		panic("slice bounds out of range")
	}

	ss := make([]*vfs.Track, len(t.s)+len(tracks))
	copy(ss, t.s[:i])
	copy(ss[i:], tracks)
	copy(ss[i+len(tracks):], t.s[i:])

	return &Tracks{s: ss}
}

func (t *Tracks) Append(tracks ...*vfs.Track) *Tracks {
	ss := make([]*vfs.Track, len(t.s)+len(tracks))
	copy(ss, t.s)
	copy(ss[len(t.s):], tracks)

	return &Tracks{s: ss}
}

func (t *Tracks) Remove(i int) *Tracks {
	ss := make([]*vfs.Track, len(t.s)-1)
	copy(ss, t.s[:i])
	copy(ss[i:], t.s[i+1:])

	return &Tracks{s: ss}
}

func (t *Tracks) RemoveRange(i int, j int) *Tracks {
	ss := make([]*vfs.Track, len(t.s)-(j-i))
	copy(ss, t.s[:i])
	copy(ss[i:], t.s[j:])

	return &Tracks{s: ss}
}
