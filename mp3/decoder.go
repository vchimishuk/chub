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

package mp3

// #cgo LDFLAGS: -lm -lmad
// #include <stdlib.h>
// #include "mp3.h"
import "C"

import (
	"errors"
	"unsafe"
)

type Decoder struct {
	cDecoder *C.struct_mp3_decoder
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Open(file string) error {
	f := C.CString(file)
	defer C.free(unsafe.Pointer(f))

	d.cDecoder = C.mp3_open(f)
	if d.cDecoder == nil {
		// TODO: Return appropriate erros. E.g. "no such file" from C.
		return errors.New("mp3_open() failed")
	}

	return nil
}

func (d *Decoder) Length() int {
	return int(d.cDecoder.length)
}

func (d *Decoder) Read(buf []byte) (read int, err error) {
	bp := (*_Ctype_char)(unsafe.Pointer(&buf[0]))
	len := (C.size_t)(len(buf))

	// TODO: Error handling.
	return int(C.mp3_decode(d.cDecoder, bp, len)), nil
}

func (d *Decoder) Seek(pos int, rel bool) {
	C.mp3_seek(d.cDecoder, C.int(pos), C.int(bool2int(rel)))
}

func (d *Decoder) Time() int {
	return int(d.cDecoder.position)
}

func (d *Decoder) Close() {
	C.mp3_close(d.cDecoder)
}

func bool2int(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}
