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

var formats map[string]Format = make(map[string]Format)

type Format interface {
	// Extensions returns a list of file extensions which belongs to the
	// format.
	Extensions() []string
	// Length returns duration of the given audio file in seconds.
	Length(file string) int
	// Tag extracts tags from audio file.
	Tag(file string) (*Tag, error)
}

func RegisterFormat(format Format) {
	for _, ext := range format.Extensions() {
		formats[ext] = format
	}
}

func format(ext string) Format {
	return formats[ext]
}
