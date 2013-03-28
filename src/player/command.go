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

	// Play given track with *vfs* playlist.
	// TODO:
	// 1. Clear *vfs* playlist.
	// 2. Add track's parent folder into *vfs*.
	// 3. Start playing *vfs* from the track but not playlist beginning.
	CommandPlayTrack
)
