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
