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
	"testing"

	"github.com/vchimishuk/chub/vfs"
)

func TestInsert(t *testing.T) {
	a := &Tracks{}

	b := a.Append(&vfs.Track{Length: 1})
	c := b.Insert(0, &vfs.Track{Length: 2})
	d := c.Insert(1, &vfs.Track{Length: 3}, &vfs.Track{Length: 4})
	e := d.Insert(3, &vfs.Track{Length: 5})
	assertTracks(t, b, 1)
	assertTracks(t, c, 2, 1)
	assertTracks(t, d, 2, 3, 4, 1)
	assertTracks(t, e, 2, 3, 4, 5, 1)
}

func TestAppend(t *testing.T) {
	a := &Tracks{}
	b := a.Append(&vfs.Track{Length: 1})
	c := b.Append(&vfs.Track{Length: 2}, &vfs.Track{Length: 3}, &vfs.Track{Length: 4})
	assertTracks(t, a)
	assertTracks(t, b, 1)
	assertTracks(t, c, 1, 2, 3, 4)
}

func TestSet(t *testing.T) {
	a := (&Tracks{}).Append(&vfs.Track{Length: 1}).Append(&vfs.Track{Length: 2})
	b := a.Set(0, &vfs.Track{Length: 3}).Set(1, &vfs.Track{Length: 4})
	assertTracks(t, b, 3, 4)
}

func TestRemove(t *testing.T) {
	a := &Tracks{}
	a = a.Append(&vfs.Track{Length: 1})
	a = a.Append(&vfs.Track{Length: 2})
	a = a.Append(&vfs.Track{Length: 3})
	b := a.Remove(1)
	assertTracks(t, b, 1, 3)
	b = b.Remove(1)
	assertTracks(t, b, 1)
	b = b.Remove(0)
	assertTracks(t, b)
}

func TestRemoveRange(t *testing.T) {
	a := &Tracks{}
	for i := 0; i < 10; i++ {
		a = a.Append(&vfs.Track{Length: i})
	}
	b := a.RemoveRange(1, 9)
	assertTracks(t, b, 0, 9)
	b = b.RemoveRange(0, 2)
	assertTracks(t, b)
}

func assertLen(t *testing.T, ts *Tracks, n int) {
	if ts.Len() != n {
		t.Fatalf("Len() %d expected got %d", n, ts.Len())
	}
}

func assertTracks(t *testing.T, ts *Tracks, vals ...int) {
	assertLen(t, ts, len(vals))

	for i := 0; i < ts.Len(); i++ {
		l := ts.Get(i).Length
		v := vals[i]
		if l != v {
			t.Fatalf("%d expected got %d", v, l)
		}
	}
}
