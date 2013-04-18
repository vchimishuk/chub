package mp3

import (
	"../vfs"
	"./id3tag"
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
type tagReaderFactory struct {
}

func (factory *tagReaderFactory) Match(file *vfs.Path) bool {
	return match(file)
}

func (factory *tagReaderFactory) TagReader() vfs.TagReader {
	return new(TagReader)
}

var TagReaderFactory vfs.TagReaderFactory = new(tagReaderFactory)
