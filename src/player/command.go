// player package is the core of the program: it manages playlists and player's state.
package player

// Available Player command code constants.
const (
	// Pause/Unpase playback.
	CommandPause int = iota
	// Start playing.
	CommandPlay
	// Stop playing process.
	CommandStop

	// Returns list with all playlists in the system.
	CommandPlaylistsList

	// Creates new playlist.
	CommandPlaylistAdd
	// Removes particular playlist.
	CommandPlaylistDelete
	// Returns playlist information.
	CommandPlaylistInfo
	// Set new name to the playlis.
	CommandPlaylistRename
)
