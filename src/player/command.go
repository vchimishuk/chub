// player package is the core of the program: it manages playlists and player's state.
package player

// Available Player command code constants.
const (
	// Pause/Unpase playback.
	CommandPause int = iota
	// Start playing if stopped, resume pause if paused
	// or start playing track from the beginning otherwise.
	CommandPlay
	// Start playing track in *vfs* playlist.
	CommandPlayTrack
	// Stop playing process.
	CommandStop

	// Returns list with all playlists in the system.
	CommandPlaylistsList

	// Creates new playlist.
	CommandPlaylistAdd
	// Append given folder to the playlist.
	CommandPlaylistAppendPath
	// Append given track to the playlist.
	CommandPlaylistAppendTrack
	// Removes particular playlist.
	CommandPlaylistDelete
	// Returns playlist information.
	CommandPlaylistInfo
	// Set new name to the playlis.
	CommandPlaylistRename
)
