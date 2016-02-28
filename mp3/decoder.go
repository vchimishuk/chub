// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of Chub.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub.  If not, see <http://www.gnu.org/licenses/>.

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
