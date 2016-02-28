package libvorbis

import "C"

// The Info structure contains basic information about the audio in a vorbis bitstream.
type Info struct {
	Version        int
	Channels       int
	Rate           int32
	BitrateUpper   int32
	BitrateNominal int32
	BitrateLower   int32
	BitrateWindow  int32
}
