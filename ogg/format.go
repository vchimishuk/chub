package ogg

import "chub/vfs"

var Format Format

type format struct {
}

func (f format) Extensions() []string {
	return []string{"ogg", "oga"}
}

func (f format) Length(file string) int {
	// TODO:
	return nil
}

func (f format) Tag(file string) (*vfs.Tag, error) {
	// TODO:
	return nil, nil
}

// ------------------------------------
/*
func tagReader() (vfs.Tag, error) {
	file, err := libvorbis.New(reader.file)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tag = vfs.Tag{}

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

func init() {
	vfs.RegisterTagReaderFactory("ogg", tagReader)
	vfs.RegisterTagReaderFactory("oga", tagReader)
}
*/
