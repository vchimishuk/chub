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

package id3tag

import (
	"testing"
)

func TestParse(t *testing.T) {
	const (
		exampleFilename = "track.mp3"
		artistExpected  = "AC/DC"
		albumExpected   = "Let There Be Rock"
		titleExpected   = "Go Down"
		numberExpected  = "01"
		yearExpected    = "1977"
	)

	tag, err := Parse(exampleFilename)
	if err != nil {
		t.Fatalf("Parsing file failed. %s", err)
	}

	if tag.Artist() != artistExpected {
		t.Fatalf("Artist() failed. Recieved '%s' but '%s' expected",
			tag.Artist(), artistExpected)
	}

	if tag.Album() != albumExpected {
		t.Fatalf("Album() failed. Recieved '%s' but '%s' expected",
			tag.Album(), albumExpected)
	}

	if tag.Title() != titleExpected {
		t.Fatalf("Title() failed. Recieved '%s' but '%s' expected",
			tag.Title(), titleExpected)
	}

	if tag.Number() != numberExpected {
		t.Fatalf("Number() failed. Recieved '%s' but '%s' expected",
			tag.Number(), numberExpected)
	}

	if tag.Year() != yearExpected {
		t.Fatalf("Year() failed. Recieved '%s' but '%s' expected",
			tag.Year(), yearExpected)
	}
}
