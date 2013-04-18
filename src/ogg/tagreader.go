// ogg TagReader implementation.
package ogg

import (
	"../vfs"
	"./libvorbis"
	"strconv"
	"strings"
)

// Ogg TagReader implementation.
type TagReader struct {
}

func (tr *TagReader) Parse(f *vfs.Path) (tag *vfs.Tag, err error) {
	file, err := libvorbis.New(f.OsPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tag = new(vfs.Tag)

	for _, uc := range file.Comment().UserComments {
		foo := strings.SplitN(uc, "=", 2)
		key := foo[0]
		value := foo[1]

		switch key {
		case libvorbis.CommentArtist:
			tag.Artist = value
		case libvorbis.CommentAlbum:
			tag.Album = value
		case libvorbis.CommentTitle:
			tag.Title = value
		case libvorbis.CommentTrackNumber:
			i, err := strconv.Atoi(value)
			if err == nil {
				tag.Number = i
			}
		}
	}

	tag.Length = 0

	return tag, nil
}

// Ogg TagReaderFactory implementation.
type tagReaderFactory struct {
}

func (factory *tagReaderFactory) Match(file *vfs.Path) bool {
	ext := strings.ToLower(file.Ext())

	// XXX: Can we support .ogv to play its sound?
	return ext == ".ogg" || ext == ".oga"
}

func (factory *tagReaderFactory) TagReader() vfs.TagReader {
	return new(TagReader)
}

var TagReaderFactory vfs.TagReaderFactory = new(tagReaderFactory)
