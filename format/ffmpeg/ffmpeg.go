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
// #cgo pkg-config: libavcodec libavformat libavutil libswresample
// #include "ffmpeg.h"
import "C"

import (
	"errors"
	"strconv"
	"unsafe"

	"github.com/vchimishuk/chub/format"
)

type metadata struct {
	artist string
	album  string
	year   int
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

func (m *metadata) Year() int {
	return m.year
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

	file := C.ffmpeg_alloc()
	err := C.ffmpeg_open(file, p)
	if err < 0 {
		C.ffmpeg_free(file)

		return nil, newError(int(err))
	}

	err = C.ffmpeg_open_codec(file)
	if err < 0 {
		C.ffmpeg_close(file)
		C.ffmpeg_free(file)

		return nil, newError(int(err))
	}

	return &decoder{file: file}, nil
}

func (d *decoder) Read(buf []byte) (read int, err error) {
	p := (*C.char)(unsafe.Pointer(&buf[0]))
	len := (C.int)(len(buf))

	n := int(C.ffmpeg_read(d.file, p, len))
	if n < 0 {
		return 0, newError(int(n))
	}

	return n, nil
}

func (d *decoder) Seek(pos int) error {
	if pos < 0 {
		return errors.New("invalid seek position")
	}
	e := C.ffmpeg_seek(d.file, C.int(pos))
	if e < 0 {
		return newError(int(e))
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

func (d *decoder) Close() error {
	C.ffmpeg_close(d.file)
	C.ffmpeg_free(d.file)

	// TODO: Return error.
	return nil
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
		"wv",
	}
}

func (f ffmpeg) Metadata(path string) (format.Metadata, error) {
	p := C.CString(path)
	defer C.free(unsafe.Pointer(p))

	file := C.ffmpeg_alloc()
	defer C.ffmpeg_free(file)

	err := C.ffmpeg_open(file, p)
	if err < 0 {
		return nil, newError(int(err))
	}
	defer C.ffmpeg_close(file)

	md := C.ffmpeg_metadata(file)
	defer C.ffmpeg_metadata_free(md)

	n, _ := strconv.Atoi(C.GoString(md.number))
	m := &metadata{
		artist: C.GoString(md.artist),
		album:  C.GoString(md.album),
		title:  C.GoString(md.title),
		year:   int(md.year),
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

func newError(err int) error {
	s := C.ffmpeg_strerror(C.int(err))
	g := C.GoString(s)
	C.free(unsafe.Pointer(s))

	return errors.New(g)
}
