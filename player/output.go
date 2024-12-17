// Copyright 2016-2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

// Output interface represents audio autput driver (ALSA, OSS, ...).
type Output interface {
	// Open opens output audio device.
	Open() error
	// Set new value for sample rate parameter.
	// TODO: Handle error by client.
	SetSampleRate(rate int) error
	// Set number of channels.
	// TODO: Handle error by client.
	SetChannels(channels int) error
	// Write new portion of data into buffer.
	Write(buf []byte) (written int, err error)
	// Reset empties ouput buffer.
	Flush() error
	// Pause pauses or resumes playback process.
	Pause() error
	// Close closes output audio device.
	Close() error
}
