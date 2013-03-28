// audio package declares some audio processing struff interfaces.
package audio

import (
	"../config"
	"fmt"
)

// Output interface represents audio autput interface (ALSA, OSS, ...).
type Output interface {
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

// OutputFactory is a creator of outputs for the particular format.
type OutputFactory interface {
	// Returns new Output driver.
	Output() (output Output, err error)
}

// List of output factories.
var outputFactories map[string]OutputFactory = make(map[string]OutputFactory)

// RegisterOutputFactory registers new output factory.
func RegisterOutputFactory(name string, factory OutputFactory) {
	outputFactories[name] = factory
}

// GetOutput returns default output driver which should be used to
// for playback.
func GetOutput() (output Output, err error) {
	name := config.Configurations.OutputName()
	factory, ok := outputFactories[name]
	if !ok {
		return nil, fmt.Errorf("Output %s not found.", name)
	}

	return factory.Output()
}

func Foo() string {
	return config.Configurations.OutputName()
}
