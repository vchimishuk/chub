// libvorbis is a libvorbis Go wrapper.
package libvorbis

// See native C API documentation: http://www.xiph.org/vorbis/doc/vorbisfile

// #cgo LDFLAGS: -lvorbisfile -lvorbis -lm -logg
// #include <stdio.h>
// #include <stdlib.h>
// #include <vorbis/codec.h>
// #include <vorbis/vorbisfile.h>
// #include "comment_hlp.h"
// #include "ogg_hlp.h"
import "C"

import (
	"errors"
	"unsafe"
)

const (
	LittleEndian = 0
	BigEndian    = 1
)

// The File structure defines an Ogg Vorbis file. 
type File struct {
	cOggFile C.OggVorbis_File
	// Specifies big or little endian byte packing.
	Endianness int
	// Specifies word size. Possible arguments are 1 for 8-bit samples, or 2 or 16-bit samples. Typical value is 2.
	WordSize int
	// Signed or unsigned data. 0 for unsigned, 1 for signed. Typically 1.
	Signed bool
}

// New is the simplest function used to open and initialize an File structure.
// It sets up all the related decoding structure. 
func New(filename string) (file *File, err error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	file = new(File)
	// Set default values.
	file.Endianness = LittleEndian
	file.WordSize = 2
	file.Signed = true

	r := C.ov_fopen(cFilename, &(file.cOggFile))
	if r != 0 {
		return nil, errors.New("Failed to open file")
	}

	return file, nil
}

// Comment returns Comment structure for the file.
func (file *File) Comment() *Comment {
	cComment := C.ov_comment(&(file.cOggFile), -1)

	comment := new(Comment)
	comment.UserComments = make([]string, cComment.comments)
	for i := 0; i < int(cComment.comments); i++ {
		cUc := C.comment_hlp_get_user_comment(cComment, _Ctype_int(i))
		comment.UserComments[i] = C.GoString(cUc)
	}
	comment.Vendor = C.GoString(cComment.vendor)

	return comment
}

// Info returns Info structure for the file.
func (file *File) Info() *Info {
	cInfo := C.ov_info(&(file.cOggFile), -1)

	info := new(Info)
	info.Version = int(cInfo.version)
	info.Channels = int(cInfo.channels)
	info.Rate = int32(cInfo.rate)
	info.BitrateUpper = int32(cInfo.bitrate_upper)
	info.BitrateNominal = int32(cInfo.bitrate_nominal)
	info.BitrateLower = int32(cInfo.bitrate_lower)
	info.BitrateWindow = int32(cInfo.bitrate_window)

	return info
}

// TimeTotal returns the total time in seconds of the physical bitstream.
func (file *File) TimeTotal() float64 {
	return float64(C.ov_time_total(&(file.cOggFile), -1))
}

// TimeTell returns the current decoding offset in seconds. 
func (file *File) TimeTell() float64 {
	return float64(C.ov_time_tell(&(file.cOggFile)))
}

// Read returns up to the specified number of bytes of decoded PCM audio.
// Return number of read 16-bit words.
func (file *File) Read(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}

	var signed int
	if file.Signed {
		signed = 1
	} else {
		signed = 0
	}

	bufLen := (C.size_t)(len(buf))
	bp := (*_Ctype_char)(unsafe.Pointer(&buf[0]))
	read := C.ogg_hlp_read(&(file.cOggFile), bp, bufLen,
		_Ctype_int(file.Endianness),
		_Ctype_int(file.WordSize),
		_Ctype_int(signed))

	return int(read)
}

// Close release file  related resources.
func (file *File) Close() {
	C.ov_clear(&(file.cOggFile))
}
