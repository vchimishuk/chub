// playlist package contains playlist managing tools and Playlist object itself.
package playlist

import (
	"errors"
)

// Playlist structure.
type Playlist struct {
	// Name of the playlist.
	name string
}

// Returns new playlist.
func New(name string) *Playlist {
	return &Playlist{name: name}
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

// Len returns number of tracks in the playlist.
func (pl *Playlist) Len() int {
	return 0 // TODO:
}
