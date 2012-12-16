// Audio output driver interface.
package audio

// Output interface represents audio autput interface (ALSA, OSS, ...).
type Output interface {
	// Returns output driver name ("alsa", "oss", etc.).
	Name() string
	// Open opens output audio device.
	Open() error
	// Set new value for sample rate parameter.
	SetSampleRate(rate int)
	// Set number of channels.
	SetChannels(channels int)
	// Wait waits some free space in output buffer, but not more than maxDelay milliseconds.
	// true result value means that output ready for new portion of data, false -- timeout has occured.
	Wait(maxDelay int) bool
	// AvailUpdate returns free size of output buffer. In bytes.
	AvailUpdate() (size int, err error)
	// Write new portion of data into buffer.
	Write(buf []byte) (written int, err error)
	// Pause pauses playback process.
	Pause()
	// Unpause release pause on playback process.
	Unpause()
	// Close closes output audio device.
	Close()
}
