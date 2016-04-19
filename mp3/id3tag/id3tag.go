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

// id3tag package implements ID3 tag parsing functionality.
// Actually this is a wrapper around C libid3tag.
package id3tag

// #cgo LDFLAGS: -lid3tag -lz
// #include <id3tag.h>
// #include <stdlib.h>
// #include "id3_hlp.h"
import "C"

import "unsafe"

// Tag struct incapulate parsed file's metadata.
type Tag struct {
	frames map[string]string
}

// Parse returns filled Tag object for the given music file.
func Parse(filename string) (tag Tag, err error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Open file.
	cId3File, err := C.id3_file_open(cFilename, C.ID3_FILE_MODE_READONLY)
	if err != nil {
		return Tag{}, err
	}
	// TODO: Can't find out why we have segfault sometimes on this close.
	//       But it seems that without this close file descriptors do not
	//       leak. So, we can live without it.
	// defer C.id3_file_close(cId3File)

	// Read tag.
	cTag, err := C.id3_file_tag(cId3File)
	if err != nil {
		return Tag{}, err
	}

	tag = Tag{frames: make(map[string]string)}

	// Parse all frames.
	for i := C.uint(0); i < cTag.nframes; i++ {
		cFrame := C.id3_hlp_get_tag_frame(cTag, i)
		cId := C.id3_hlp_get_frame_id(cFrame)
		// XXX: As I understood cId memory will be GCed with id
		//      (they share same memory).
		id := C.GoString(cId)
		cVal := C.id3_hlp_get_frame_string(cFrame)
		val := C.GoString(cVal)

		tag.frames[id] = val
	}

	return tag, nil
}

// Artist returns name of the artist.
func (tag Tag) Artist() string {
	return tag.frames["TPE1"]
}

// Album returns name of the album.
func (tag Tag) Album() string {
	return tag.frames["TALB"]
}

// Artist returns track's title.
func (tag Tag) Title() string {
	return tag.frames["TIT2"]
}

// Number returns track's number.
func (tag Tag) Number() string {
	return tag.frames["TRCK"]
}

// Year returns track's year.
func (tag Tag) Year() string {
	year, present := tag.frames["TDRC"]
	if !present {
		year, _ = tag.frames["TYER"]
	}

	return year
}

// Comment returns track's comment string.
//func (tag *Tag) Comment() string {
//	fmt.Printf("%v\n", tag.frames)
//	return tag.frames["TCON"]
//}
