package mp3

import (
	"../audio"
	"../vfs"
	"./libmad"
)

// mp3 decoder implementation.
type decoder struct {
	mad *libmad.Decoder
}

func newDecoder(track *vfs.Track) (d audio.Decoder, err error) {
	d = new(decoder)

	return d, nil
}

func (d *decoder) Match(track *vfs.Track) bool {
	return match(track.Path)
}

func (d *decoder) Open(track *vfs.Track) error {
	mad, err := libmad.New(track.Path.OsPath())
	d.mad = mad

	return err
}

func (d *decoder) Read(buf []byte) (read int, err error) {
	// TODO: Maybe improve libmad to return error with Read.
	read = d.mad.Read(buf)

	return read, nil
}

func (d *decoder) Close() {
	d.mad.Close()
}

type decoderFactory struct {
}

func (factory *decoderFactory) Match(track *vfs.Track) bool {
	return match(track.Path)
}

func (factory *decoderFactory) Decoder(track *vfs.Track) (decoder audio.Decoder, err error) {
	return newDecoder(track)
}

var DecoderFactory audio.DecoderFactory = new(decoderFactory)
