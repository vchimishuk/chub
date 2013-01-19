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
