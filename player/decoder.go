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

package player

// Decoder interface represents audio decoder for the particular audio format.
type Decoder interface {
	// Open and prepare file for decoding.
	Open(file string) error
	// Read decode piece of data and returns raw PCM audio data.
	Read(buf []byte) (read int, err error)
	// Time returns current decoded position in seconds.
	Time() int
	// Close releases decoder resources.
	Close()
}