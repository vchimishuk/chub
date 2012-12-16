// player package is the core of the program: it manages playlists and player's state.
package player

// Playlist structure.
type Playlist struct {
	// Name of the playlist.
	Name string
	// Defines is playlist is system or not.
	// System playlists can't be deleted or edited by user, they are
	// totaly managed by system. Now we have only one system playlist *vfs*
	// which represents directory of the current playing file.
	system bool
}

// Returns new playlist.
func newPlaylist(name string) *Playlist {
	return &Playlist{Name: name}
}
