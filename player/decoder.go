package player

// Decoder interface represents audio decoder for the particular audio format.
type Decoder interface {
	// Open and prepare file for decoding.
	Open(file string) error
	// Read decode piece of data and returns raw PCM audio data.
	Read(buf []byte) (read int, err error)
	// Close releases decoder resources.
	Close()
}
