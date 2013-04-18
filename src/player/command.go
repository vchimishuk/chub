// player package is the core of the program: it manages playlists and player's state.
package player

// Available commandRoutine command code constants.
const (
	// Pause/Unpase playback.
	commandPause int = iota
	// Start playing if stopped, resume pause if paused
	// or start playing track from the beginning otherwise.
	commandPlay
	commandPlayTrack
	commandPlayPlaylist
	// Stop playing process.
	commandStop

	commandPlaylists

	commandAddPlaylist
	commandAppendTrack
	commandDeletePlaylist
	// Returns playlist information.
	commandPlaylistInfo
	// Set new name to the playlis.
	commandPlaylistRename
)

// Available playingRoutine commands.
const (
	playingCommandStop int = iota
	playingCommandPause
)
