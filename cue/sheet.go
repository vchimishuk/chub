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

package cue

// Cue sheet file representation.
type Sheet struct {
	// Disc's media catalog number.
	Catalog string
	// Name of a perfomer for a CD-TEXT enhanced disc
	Performer string
	// Specify a title for a CD-TEXT enhanced disc.
	Title string
	// Specify songwriter for disc.
	Songwriter string
	// Comments in the CUE SHEET file.
	Comments []string
	// Name of the file that contains the encoded CD-TEXT information for the disc.
	CdTextFile string
	// Data/audio files descibed byt the cue-file.
	Files []*File
}

// Type of the audio file.
type FileType int

const (
	// Intel binary file (least significant byte first)
	FileTypeBinary FileType = iota
	// Motorola binary file (most significant byte first)
	FileTypeMotorola
	// Audio AIFF file
	FileTypeAiff
	// Audio WAVE file
	FileTypeWave
	// Audio MP3 file
	FileTypeMp3
)

// Track datatype.
type TrackDataType int

const (
	// AUDIO – Audio/Music (2352)
	DataTypeAudio = iota
	// CDG – Karaoke CD+G (2448)
	DataTypeCdg
	// MODE1/2048 – CDROM Mode1 Data (cooked)
	DataTypeMode1_2048
	// MODE1/2352 – CDROM Mode1 Data (raw)
	DataTypeMode1_2352
	// MODE2/2336 – CDROM-XA Mode2 Data
	DataTypeMode2_2336
	// MODE2/2352 – CDROM-XA Mode2 Data
	DataTypeMode2_2352
	// CDI/2336 – CDI Mode2 Data
	DataTypeCdi_2336
	// CDI/2352 – CDI Mode2 Data
	DataTypeCdi_2352
)

// Time point description type.
type Time struct {
	// Minutes.
	Min int
	// Minutes.
	Sec int
	// Frames.
	Frames int
}

// Seconds returns length in seconds.
func (time *Time) Seconds() int {
	return time.Min*60 + time.Sec
}

// Track index type
type Index struct {
	// Index number.
	Number int
	// Index starting time.
	Time *Time
}

// Additional decode information about track.
type TrackFlag int

const (
	// Digital copy permitted.
	TrackFlagDcp = iota
	// Four channel audio.
	TrackFlag4ch
	// Pre-emphasis enabled (audio tracks only).
	TrackFlagPre
	// Serial copy management system (not supported by all recorders).
	TrackFlagScms
)

type Track struct {
	// Track number (1-99).
	Number int
	// Track datatype.
	DataType TrackDataType
	// Track title.
	Title string
	// Track preformer.
	Performer string
	// Songwriter.
	Songwriter string
	// Track decode flags.
	Flags []TrackFlag
	// Internetional Standaard Recording Code.
	Isrc string
	// Track indexes.
	Indexes []*Index
	// Length of the track pregap.
	Pregap *Time
	// Length of the track postgap.
	Postgap *Time
}

// Audio file representation structure.
type File struct {
	// Name (path) of the file.
	Name string
	// Type of the audio file.
	Type FileType
	// List of present tracks in the file.
	Tracks []*Track
}
