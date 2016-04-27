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

// Decoder interface represents audio decoder for the particular audio format.
type Decoder interface {
	// Open and prepare file for decoding.
	Open(file string) error
	// Read decode piece of data and returns raw PCM audio data.
	Read(buf []byte) (read int, err error)
	// Seek sets new position in seconds to start decoding from. If rel
	// parameter is set new position will be calculated as current plus
	// pos, otherwise pos is an absolute seconds position.
	Seek(pos int, rel bool)
	// Time returns current decoded position in seconds.
	Time() int
	// SampleRate returns sample rate of decoded stream.
	SampleRate() int
	// Channels returns number of channels in decoded stream.
	Channels() int
	// Close releases decoder resources.
	Close()
}
