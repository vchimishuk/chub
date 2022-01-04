// Copyright 2019 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

// FFmpeg audio decoder format.
package ffmpeg

// #cgo CFLAGS: -Wno-deprecated-declarations
// #cgo LDFLAGS: -lavcodec -lavformat -lavutil -lswresample
// #include "ffmpeg.h"
import "C"

import (
	"errors"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/vchimishuk/chub/format"
)

type metadata struct {
	artist string
	album  string
	title  string
	number int
	length int
}

func (m *metadata) Artist() string {
	return m.artist
}

func (m *metadata) Album() string {
	return m.album
}

func (m *metadata) Title() string {
	return m.title
}

func (m *metadata) Number() int {
	return m.number
}

func (m *metadata) Length() int {
	return m.length
}

type decoder struct {
	file *C.struct_ffmpeg_file
}

func newDecoder(path string) (*decoder, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))

	file := C.ffmpeg_open(p)
	if file == nil {
		// TODO: Handle error.
		panic("TODO: ")
	}
	e := C.ffmpeg_open_codec(file)
	if e != 0 {
		C.ffmpeg_close(file)
		return nil, errors.New("TODO:")
	}

	return &decoder{file: file}, nil
}

func (d *decoder) Read(buf []byte) (read int, err error) {
	p := (*C.char)(unsafe.Pointer(&buf[0]))
	len := (C.int)(len(buf))

	n := int(C.ffmpeg_read(d.file, p, len))
	if n < 0 {
		return 0, fmt.Errorf("TODO: %d", n)
	}

	return n, nil
}

func (d *decoder) Seek(pos int, rel bool) error {
	if !rel && pos < 0 {
		return errors.New("invalid seek position")
	}
	e := C.ffmpeg_seek(d.file, C.int(pos), C.int(btoi(rel)))
	if e < 0 {
		return errors.New("TODO:")
	}

	return nil
}

func (d *decoder) Time() int {
	return int(C.ffmpeg_time(d.file))
}

func (d *decoder) SampleRate() int {
	return int(C.ffmpeg_sample_rate(d.file))
}

func (d *decoder) Channels() int {
	return int(C.ffmpeg_channels(d.file))
}

func (d *decoder) Close() {
	C.ffmpeg_close(d.file)
}

type ffmpeg struct {
}

func NewFormat() format.Format {
	C.ffmpeg_init()

	return &ffmpeg{}
}

func (f ffmpeg) Extensions() []string {
	return []string{
		"ape",
		"flac",
		"mp3",
		"ogg",
	}
}

func (f ffmpeg) Metadata(path string) (format.Metadata, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))

	file := C.ffmpeg_open(p)
	if file == nil {
		// TODO: Handle error.
		panic("TODO: ")
	}
	defer C.ffmpeg_close(file)
	md := C.ffmpeg_metadata(file)
	defer C.ffmpeg_metadata_free(md)

	n, _ := strconv.Atoi(C.GoString(md.number))
	m := &metadata{
		artist: C.GoString(md.artist),
		album:  C.GoString(md.album),
		title:  C.GoString(md.title),
		number: n,
		length: int(md.duration),
	}

	return m, nil
}

func (f ffmpeg) Decoder(path string) (format.Decoder, error) {
	return newDecoder(path)
}

func btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}
