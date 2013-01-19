// id3tag package implements ID3 tag parsing functionality.
// Actually this is a wrapper around C libid3tag.
package id3tag

// #cgo LDFLAGS: -lid3tag -lz
// #include <id3tag.h>
// #include <stdlib.h>
// #include "id3_hlp.h"
import "C"

import (
	"unsafe"
)

// Tag struct incapulate parsed file's metadata.
type Tag struct {
	cTag   *_Ctype_struct_id3_tag
	frames map[string]string
}

// Parse returns filled Tag object for the given music file.
func Parse(filename string) (tag *Tag, err error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Open file.
	cId3File, err := C.id3_file_open(cFilename, C.ID3_FILE_MODE_READONLY)
	if err != nil {
		return nil, err
	}
	defer C.id3_file_close(cId3File)

	tag = new(Tag)
	tag.frames = make(map[string]string)

	// Read tag.
	tag.cTag, err = C.id3_file_tag(cId3File)
	if err != nil {
		return nil, err
	}
	// Tag struct will be released with file's one.

	// Parse all frames.
	for i := _Ctype_uint(0); i < tag.cTag.nframes; i++ {
		cFrame := C.id3_hlp_get_tag_frame(tag.cTag, i)
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
func (tag *Tag) Artist() string {
	return tag.frames["TPE1"]
}

// Album returns name of the album.
func (tag *Tag) Album() string {
	return tag.frames["TALB"]
}

// Artist returns track's title.
func (tag *Tag) Title() string {
	return tag.frames["TIT2"]
}

// Number returns track's number.
func (tag *Tag) Number() string {
	return tag.frames["TRCK"]
}

// Year returns track's year.
func (tag *Tag) Year() string {
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
