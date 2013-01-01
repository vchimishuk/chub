// Virtual File System package implements abstraction layer OS file system API
// and tracks representation in the physical files.
package vfs

import (
	"sort"
)

// Directory is a filesystem entry structure which representing
// directories.
type Directory struct {
	// Path to the file.
	Path *Path
	// Directory name.
	Name string
}

// directorySlice attaches sorting methods to the []*Directory slice type.
type directorySlice []*Directory

// Len returns length of the directories slice.
func (ds directorySlice) Len() int {
	return len(ds)
}

// Swap swaps two elements in directories slice.
func (ds directorySlice) Swap(i, j int) {
	ds[i], ds[j] = ds[j], ds[i]
}

// Less returns true if [i] slice element less than [j] and whey are should be swapped
// during sorting.
func (ds directorySlice) Less(i, j int) bool {
	return ds[i].Name < ds[j].Name
}

// Sort method is a shortcut method for sort.Sort(directorySlice).
func (ds directorySlice) Sort() {
	sort.Sort(ds)
}

// Track is a filesystem entry structure representing track.
// In VFS terms track can represents whole file (usually MP3 file
// in some particular folder equals one track from an album), or piece
// of some physical file (e. g. FLAC file can consists from many tracks).
type Track struct {
	// Path to the file.
	Path *Path
	// Track media information.
	Tag int // XXX: ??? //*audio.Tag
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

// trackSlice attaches sorting methods to the []*Track slice type.
type trackSlice []*Track

// Len returns length of the tracks slice.
func (ts trackSlice) Len() int {
	return len(ts)
}

// Swap swaps two elements in tracks slice.
func (ts trackSlice) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

// Less returns true if [i] slice element less than [j] and whey are should be swapped
// during sorting.
func (ts trackSlice) Less(i, j int) bool {
	return false // TODO:
}

// Sort method is a shortcut method for sort.Sort(trackSlice).
func (ts trackSlice) Sort() {
	sort.Sort(ts)
}
