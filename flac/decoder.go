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

// #cgo LDFLAGS: -lFLAC
// #include "flac.h"
import "C"
import (
	"errors"
	"unsafe"
)

type Decoder struct {
	cDecoder *C.struct_flac_decoder
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Open(file string) error {
	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))

	d.cDecoder = C.flac_open(cFile)
	if d.cDecoder == nil {
		// TODO: Return appropriate erros. E.g. "no such file" from C.
		return errors.New("flac_open() failed")
	}

	return nil
}

// Read decodes and read next portion of data into buf.
// Returns number of bytes written in buf or 0 if end of the stream was reached.
func (d *Decoder) Read(buf []byte) (int, error) {
	bp := (*C.char)(unsafe.Pointer(&buf[0]))
	len := (C.int)(len(buf))

	// TODO: Error handling.
	return int(C.flac_decode(d.cDecoder, bp, len)), nil
}

// Time returns current decoded position in seconds.
func (d *Decoder) Time() int {
	return int(C.flac_time(d.cDecoder))
}

// Seek sets new decoding position to absolute number of seconds
// if rel parameter is set to true. If rel parameter is set to false
// new position calculated as current one plus pos.
func (d *Decoder) Seek(pos int, rel bool) {
	C.flac_seek(d.cDecoder, C.int(pos), C.int(bool2int(rel)))
}

func (d *Decoder) SampleRate() int {
	return int(C.flac_sample_rate(d.cDecoder))
}

func (d *Decoder) Channels() int {
	return int(C.flac_channels(d.cDecoder))
}

// Length returns file length in seconds.
func (d *Decoder) Length() int {
	return int(C.flac_length(d.cDecoder))
}

// Close closes decoder and frees its resources.
func (d *Decoder) Close() {
	C.flac_close(d.cDecoder)
}

func bool2int(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}
