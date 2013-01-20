// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

// Directory is a filesystem entry structure which representing
// directories.
type Directory struct {
	// Path to the file.
	Path *Path
	// Directory name.
	Name string
}

// Track's tag data.
type Tag struct {
	// Artist name.
	Artist string
	// Album name.
	Album string
	// Track's title.
	Title string
	// Track number.
	Number string
	// Track length
	Length string
}

// Track is a filesystem entry structure representing track.
// In VFS terms track can represents whole file (usually MP3 file
// in some particular folder equals one track from an album), or piece
// of some physical file (e. g. FLAC file can consists from many tracks).
type Track struct {
	// Path to the file.
	Path *Path
	// Track media information.
	Tag *Tag
	// If Part is true it means this track represents piece of the 
	// actual file, not the whole file. E. g. we have album FLAC file
	// and this track points only to one song in this FLAC file (which is
	// album itself).
	Part bool
	// Track beginning in the physical file.
	Start int
	// Track end position in the physical file.
	End int
}
