// Audio decoder interface.
package audio

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
