// playlist package contains playlist managing tools and Playlist object itself.
package playlist

import (
	"../../vfs"
	"container/list"
	"errors"
	"sync"
)

// Playlist structure.
type Playlist struct {
	// Name of the playlist.
	name string
	// Contained tracks.
	// TODO: Replace list with slice.
	tracks *list.List
	// Playlist modification lock. Only player's routine and playing thread
	// routines can manage playlists, and since playing thread routine
	// can access only current playing playlist only the last one should
	// be locked before modification. So, all other playlists don't use
	// this lock (till they become current).
	lock sync.Mutex
}

// Returns new playlist.
func New(name string) *Playlist {
	return &Playlist{name: name, tracks: list.New()}
}

// System returns true if the playlist is system.
// System playlist can't be deleted, edited or renamed by the user,
// it is totaly managed by system. Fo instance *vfs* playlist
// represents directory of the current playing file.
func (pl *Playlist) System() bool {
	l := len(pl.name)

	return l >= 2 && pl.name[0] == '*' && pl.name[l-1] == '*'
}

// Name returns name of the playlist.
func (pl *Playlist) Name() string {
	return pl.name
}

// Rename the playlist. Notice: only non system playlists can be renamed.
func (pl *Playlist) Rename(name string) error {
	if pl.System() {
		return errors.New("System playlist can't be renamed.")
	}

	pl.name = name

	return nil
}

// Get returns track by its index.
func (pl *Playlist) Get(i int) *vfs.Track {
	if i >= pl.Len() {
		panic("Track index out of range.")
	}

	entry := pl.tracks.Front()
	for i := 0; i < i; i++ {
		entry = entry.Next()
	}

	return entry.Value.(*vfs.Track)
}

// Len returns number of tracks in the playlist.
func (pl *Playlist) Len() int {
	return pl.tracks.Len()
}

// Remove all tracks from the playlist.
func (pl *Playlist) Clear() {
	pl.tracks.Init()
}

// Append tracks to the playlist.
func (pl *Playlist) Append(track ...*vfs.Track) {
	for _, t := range track {
		pl.tracks.PushBack(t)
	}
}

// Lock current playlist for modification.
func (pl *Playlist) Lock() {
	pl.lock.Lock()
}

// Release modification lock.
func (pl *Playlist) Unlock() {
	pl.lock.Unlock()
}
