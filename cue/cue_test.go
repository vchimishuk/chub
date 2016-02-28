package cue

import (
	"testing"
)

const (
	expectedTitle        = "Doro"
	expectedPerformer    = "Doro"
	expectedFilesNumber  = 1
	expectedFileName     = "Doro - Doro.ape"
	expectedTracksNumber = 10
	// Expected first track values.
	expectedTrackNumber        = 1
	expectedTrackTitle         = "Unholy Love"
	expectedTrackPerformer     = "Doro"
	expectedTrackIndexesNumber = 1
	expectedIndexNumber        = 1
	expectedTrackIndexNumber   = 1
)

func TestPackage(t *testing.T) {
	filename := "test.cue"

	sheet, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("Failed to parse file. %s", err)
	}

	if sheet.Title != expectedTitle {
		t.Fatalf("Expected title %s but %s got.",
			expectedTitle, sheet.Title)
	}
	if sheet.Performer != expectedPerformer {
		t.Fatalf("Expected performer %s but %s got.",
			expectedPerformer, sheet.Performer)
	}

	if len(sheet.Files) != expectedFilesNumber {
		t.Fatalf("Expected files number %d but %d got.",
			expectedFilesNumber, len(sheet.Files))
	}

	file := sheet.Files[0]

	if file.Name != expectedFileName {
		t.Fatalf("Expected file name %s but %s got.",
			expectedFileName, file.Name)
	}
	if len(file.Tracks) != expectedTracksNumber {
		t.Fatalf("Expected tracks number %d but %d got.",
			expectedTracksNumber, len(file.Tracks))
	}

	// Assert first track only.
	track := file.Tracks[0]
	if track.Number != expectedTrackNumber {
		t.Fatalf("Expected track number %d but %d got.",
			expectedTrackNumber, track.Number)
	}
	if track.Title != expectedTrackTitle {
		t.Fatalf("Expected track title %s but %s got.",
			expectedTrackTitle, track.Title)
	}
	if track.Performer != expectedTrackPerformer {
		t.Fatalf("Expected track performer %s but %s got.",
			expectedTrackPerformer, track.Performer)
	}
	if len(track.Indexes) != expectedTrackIndexesNumber {
		t.Fatalf("Expected track indexes number %d but %d got.",
			expectedTrackIndexesNumber, len(track.Indexes))
	}

	index := track.Indexes[0]
	if index.Number != expectedIndexNumber {
		t.Fatalf("Expected track indexes number %d but %d got.",
			expectedTrackIndexNumber, index.Number)
	}
	time := index.Time
	if time.Min != 0 || time.Sec != 0 || time.Frames != 0 {
		t.Fatalf("Expected track index time 0:0:0 but %d:%d:%d got.",
			time.Min, time.Sec, time.Frames)
	}
}
