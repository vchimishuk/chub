package mp3

import "github.com/vchimishuk/chub/mp3/libmad"

// mp3 decoder implementation.
type Decoder struct {
	mad *libmad.Decoder
}

func NewDecoder() *Decoder {
	return new(Decoder)
}

func (d *Decoder) Open(file string) error {
	mad, err := libmad.New(file)
	if err != nil {
		return err
	}
	d.mad = mad

	return nil
}

func (d *Decoder) Read(buf []byte) (read int, err error) {
	// TODO: Maybe improve libmad to return error with Read.
	read = d.mad.Read(buf)

	return read, nil
}

func (d *Decoder) Close() {
	d.mad.Close()
}
