package audio

import (
	"../vfs"
)

// Track's tag data.
type Tag struct {
	// Artist name.
	Artist string
	// Album name.
	Album string
	// Track's title.
	Title string
	// Track number.
	Number string
	// Track length
	Length string
}

// TagReader interface wraps methods for working with audio file tags.
type TagReader interface {
	// ReadTag parse audio file's metadata and returns filled Tag object.
	Parse(file *vfs.Path) (tag *Tag, err error)
}

// TagReaderFactory interface is a creator of TagReader
// objects of particular implementation.
type TagReaderFactory interface {
	// Match returns true it given file can be processed with current TagReader.
	Match(file *vfs.Path) bool
	// Returns new TagReader implementation.
	TagReader() TagReader
}

// All registered TagReader's.
var tagReaderFactories []TagReaderFactory

// NewTagReader returns TagReader implementation for the given audio file.
// If there is not TagReader that supports this filetype nil will be returned.
func NewTagReader(file *vfs.Path) TagReader {
	for _, factory := range tagReaderFactories {
		if factory.Match(file) {
			return factory.TagReader()
		}
	}

	return nil
}

// RegisterTagReaderFactory registers new TagReader factory implementation.
func RegisterTagReaderFactory(factory TagReaderFactory) {
	tagReaderFactories = append(tagReaderFactories, factory)
}
