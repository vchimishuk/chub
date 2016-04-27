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

package flac

import (
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/vfs"
)

var Format format

type format struct {
}

func (f format) Extensions() []string {
	return []string{"flac"}
}

func (f format) Length(file string) int {
	d := NewDecoder()
	defer d.Close()

	d.Open(file)
	// TODO: Check error. Add error in Decoder.Length() signature.

	return d.Length()
}

func (f format) Tag(file string) (*vfs.Tag, error) {
	// TODO:
	return &vfs.Tag{Artist: "FLAC",
			Album:  "FLAC",
			Title:  "FLAC",
			Number: 0},
		nil
}

func (f format) Decoder() player.Decoder {
	return NewDecoder()
}
