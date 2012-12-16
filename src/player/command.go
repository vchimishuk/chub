// player package is the core of the program: it manages playlists and player's state.
package player

// Available Player command code constants.
const (
	// Returns list of all loaded playlists.
	CMD_PLAYLISTS_LIST int = iota
	// Creates new playlist.
	CMD_PLAYLISTS_ADD
	// Removes particular playlist.
	CMD_PLAYLISTS_DELETE
	// Returns specified playlist tracks.
	CMD_PLAYLIST_LIST
	// Append track to the playlist.
	CMD_PLAYLIST_ADD
	// Starts playing track.
	CMD_PLAY
	// Pauses or resumes playing if any.
	CMD_PAUSE
	// Stops playing if any.
	CMD_STOP
)
