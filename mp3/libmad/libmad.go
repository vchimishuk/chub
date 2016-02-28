// libmad is MPEG Audio Decoder.
// TODO: Rename to mad.
package libmad

// #cgo LDFLAGS: -lm -lmad
// #include <stdlib.h>
// #include "libmad_hlp.h"
import "C"

import (
	"fmt"
	"unsafe"
)

// File structure represents MP3 file.
type Decoder struct {
	// C decoder structure.
	cDecoder C.struct_gomad_decoder
}

// Whence type for Seek function.
type Whence int

const (
	// Set new position relative to the start.
	SeekSet = iota
	// Set new position related to the current one.
	SeekCurrent
)

// New opens and initialize MAD decoder and File structure.
func New(filename string) (decoder *Decoder, err error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	decoder = new(Decoder)

	_, e := C.gomad_open(&(decoder.cDecoder), cFilename)
	if e != nil {
		return nil, fmt.Errorf("Failed to open file %s: %s", filename, e)
	}

	return decoder, nil
}

// SampleRate returns file's sample rate value.
func (decoder *Decoder) SampleRate() int {
	return int(decoder.cDecoder.sample_rate)
}

// Channels returns number of channels for the related audio file.
func (decoder *Decoder) Channels() int {
	return int(decoder.cDecoder.channels)
}

// Length returns file's length in seconds.
func (decoder *Decoder) Length() int {
	return int(decoder.cDecoder.length)
}

// CurrentPosition returns current decoding position, in seconds.
func (decoder *Decoder) CurrentPosition() int {
	return int(decoder.cDecoder.current_position)
}

// Seek move decoding position.
// If whence parameter is SeekSet than position parameter should be non-negative
// integer value, which is new decoding position relative to the start of the file.
// If whence equals SeekCurrent than position parameter can be negative as well as positive integer,
// which means new position related to the current one.
func (decoder *Decoder) Seek(position int, whence Whence) error {
	// TODO:

	return nil
}

// Read returns up to the specified number of bytes of decoded PCM audio.
// Return number of read 16-bit words.
func (decoder *Decoder) Read(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}

	bp := (*_Ctype_char)(unsafe.Pointer(&buf[0]))
	l := (C.size_t)(len(buf))

	read := C.gomad_read(&decoder.cDecoder, bp, l)

	return int(read)
}

// Close release resources assigned to Decoder structure and close MAD decoder.
func (decoder *Decoder) Close() {
	C.gomad_close(&(decoder.cDecoder))
}
