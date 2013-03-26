package audio

import (
	"fmt"
)

// Decoder interface represents audio decoder for the particular audio format.
type Decoder interface {
	// Match returns true if given file supported by this decoder.
	Match(filename string) bool
	// Open inialize decoder object.
	Open(filename string) error
	// Read decode piece of data and returns raw PCM audio data.
	Read(buf []byte) (read int, err error)
	// Close releases decoder resources.
	Close()
}

// DecoderFactory is a creator of decoders for the particular format.
type DecoderFactory interface {
	// Match return true if the track audio format is supported by
	// underlying decoder.
	Match(track int /* TODO: Find by track? */) bool
	// Returns new Decoder
	Decoder(track int /* TODO: */) (decoder Decoder, err error)
}

// List of decoder factories.
var decoderFactories []DecoderFactory

// RegisterDecoderFactory registers new decoder factory.
func RegisterDecoderFactory(factory DecoderFactory) {
	decoderFactories = append(decoderFactories, factory)
}

// GetDecoder returns decoder for decoding given track.
func GetDecoder(track int /* TODO: Track? */) (decoder Decoder, err error) {
	for _, factory := range decoderFactories {
		if factory.Match(track) {
			return factory.Decoder(track)
		}
	}

	return nil, fmt.Errorf("Decoder not found for file '%s'", "???") // TODO:
}
