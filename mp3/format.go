package mp3

import (
	"strconv"

	"github.com/vchimishuk/chub/mp3/id3tag"
	"github.com/vchimishuk/chub/mp3/libmad"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/vfs"
)

var Format format

type format struct {
}

func (f format) Extensions() []string {
	return []string{"mp3"}
}

func (f format) Length(file string) int {
	m, err := libmad.New(file)
	if err != nil {
		return 0
	}
	defer m.Close()

	return m.Length()
}

func (f format) Tag(file string) (*vfs.Tag, error) {
	id3Tag, err := id3tag.Parse(file)
	if err != nil {
		return nil, err
	}
	number, err := strconv.Atoi(id3Tag.Number())
	if err != nil {
		number = 0
	}

	tag := &vfs.Tag{
		Artist: id3Tag.Artist(),
		Album:  id3Tag.Album(),
		Title:  id3Tag.Title(),
		Number: number}

	return tag, nil
}

func (f format) Decoder() player.Decoder {
	return NewDecoder()
}
