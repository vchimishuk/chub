package vfs

var formats map[string]Format = make(map[string]Format)

type Format interface {
	// Extensions returns a list of file extensions which belongs to the
	// format.
	Extensions() []string
	// Length returns duration of the given audio file in seconds.
	Length(file string) int
	// Tag extracts tags from audio file.
	Tag(file string) (*Tag, error)
}

func RegisterFormat(format Format) {
	for _, ext := range format.Extensions() {
		formats[ext] = format
	}
}

func format(ext string) Format {
	return formats[ext]
}
