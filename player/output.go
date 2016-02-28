package player

// Output interface represents audio autput driver (ALSA, OSS, ...).
type Output interface {
	// Open opens output audio device.
	Open() error
	// Set new value for sample rate parameter.
	SetSampleRate(rate int)
	// Set number of channels.
	SetChannels(channels int)
	// Wait waits some free space in output buffer, but not more than
	// maxDelay milliseconds. true result value means that output is
	// ready for new portion of data, false -- timeout has occured.
	Wait(maxDelay int) (ok bool, err error)
	// AvailUpdate returns free size of output buffer. In bytes.
	AvailUpdate() (size int, err error)
	// Write new portion of data into buffer.
	Write(buf []byte) (written int, err error)
	// Pause pauses or resumes playback process.
	Pause()
	// Paused returns true if output driver is in paused state now.
	Paused() bool
	// Close closes output audio device.
	Close()
}
