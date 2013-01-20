// mp3 TagReader implementation.
package mp3

import (
	"../vfs"
	"./id3tag"
	"strings"
)

// MP3 TagReader implementation.
type TagReader struct {
}

func (tr *TagReader) Parse(file *vfs.Path) (tag *vfs.Tag, err error) {
	id3Tag, err := id3tag.Parse(file.OsPath())
	if err != nil {
		return nil, err
	}

	tag = new(vfs.Tag)
	tag.Artist = id3Tag.Artist()
	tag.Album = id3Tag.Album()
	tag.Title = id3Tag.Title()
	tag.Length = 0

	return tag, nil
}

// MP3 TagReaderFactory implementation.
type TagReaderFactory struct {
}

func (factory *TagReaderFactory) Match(file *vfs.Path) bool {
	ext := strings.ToLower(file.Ext())

	return ext == ".mp3"
}

func (factory *TagReaderFactory) TagReader() vfs.TagReader {
	return new(TagReader)
}

// Init is a dummy function and used in the main source file only to make
// this package loads.
func Init() {

}

// Register this implementation.
func init() {
	vfs.RegisterTagReaderFactory(new(TagReaderFactory))
}
