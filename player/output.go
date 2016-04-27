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

// Output interface represents audio autput driver (ALSA, OSS, ...).
type Output interface {
	// Open opens output audio device.
	Open() error
	// Returns current sample rate.
	SampleRate() int
	// Set new value for sample rate parameter.
	SetSampleRate(rate int)
	// Returns current channels value.
	Channels() int
	// Set number of channels.
	SetChannels(channels int)
	// Wait waits some free space in output buffer, but not more than
	// maxDelay milliseconds. true result value means that output is
	// ready for new portion of data, false -- timeout has occured.
	Wait(maxDelay int) (ok bool, err error)
	// AvailUpdate returns free size of output buffer. In bytes.
	AvailUpdate() (size int, err error)
	// Write new portion of data into buffer.
	Write(buf []byte) (written int, err error)
	// Pause pauses or resumes playback process.
	Pause()
	// Paused returns true if output driver is in paused state now.
	Paused() bool
	// Close closes output audio device.
	Close()
}
