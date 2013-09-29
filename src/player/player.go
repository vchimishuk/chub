// player package is the core of the program: it manages playlists and player's state.
package player

import (
	"../vfs"
	"./playingthread"
	"./playlist"
	"errors"
	"fmt"
)

const (
	PlaylistVfs = "*vfs*"
)

const (
	stateStopped int = iota
	statePlaying
	statePaused
)

// State listener can be attached to the player for be notified
// for player state change events.
type StateListener interface {
	// Playlist with name "name" was changed, e.g. track added or deleted.
	PlaylistChanged(name string)
	// Playlists list was changed (playlist was added, removed, etc).
	PlaylistsChanged()
	// Player's state is changed, e.g player was paused or stopped.
	StateChanged() // TODO: Add new state parameter.
	// Current playing track was changed.
	TrackChanged() // TODO: Add new track parameter.
	// Volume level changed.
	VolumeChanged() // TODO: Add volume value parameter.
}

// Player engine object.
type Player struct {
	// All (user and system) playlists list.
	playlists []*playlist.Playlist
	// Playlist which are playing now.
	currentPlaylist *playlist.Playlist
	// Current playing track if any.
	currentTrack *vfs.Track
	// Channel to communicate player with. Client code can
	// write commands and read responses to/from the channel.
	commandChan chan *command
	// Playing thread, -- the core decode -> output object.
	playingThread *playingthread.PlayingThread
	// Listener to be notified about player state changes.
	stateListener StateListener
}

// New returns a newly created Player object.
func New() *Player {
	p := new(Player)
	p.commandChan = make(chan *command, 10)
	// Create some predefined system playlists.
	p.playlists = append(p.playlists, playlist.New(PlaylistVfs))

	return p
}

// SetStateListener attaches new state listener to the player.
func (player *Player) SetStateListener(l StateListener) {
	player.stateListener = l
}

// Run starts Player execution.
func (player *Player) Run() error {
	go player.commandRoutine()

	return nil
}

// PlayTrack starts playing track in *vfs* playlist.
func (player *Player) PlayTrack(track *vfs.Track) error {
	cmd := newCommand(player.commandPlayTrack, track)
	res := player.commandDispatcher(cmd)

	return res.err
}

// PlayPlaylist starts playing given playlist from the given position in the playlist.
// First track in the playlist has 0 position.
func (player *Player) PlayPlaylist(name string, track int) error {
	cmd := newCommand(player.commandPlayPlaylist, name, track)
	res := player.commandDispatcher(cmd)

	return res.err
}

// Playlists eturns list with all playlists in the system.
func (player *Player) Playlists() []*playlist.Playlist {
	cmd := newCommand(player.commandPlaylists)
	res := player.commandDispatcher(cmd)
	plists := res.args[0].([]*playlist.Playlist)

	return plists
}

// Playlist returns playlist's content (list of containing tracks).
func (player *Player) Playlist(name string) (plist *playlist.Playlist, err error) {
	cmd := newCommand(player.commandPlaylist, name)
	res := player.commandDispatcher(cmd)

	if res.err != nil {
		return nil, res.err
	}

	return res.args[0].(*playlist.Playlist), nil
}

// AddPlaylist creates new playlist.
func (player *Player) AddPlaylist(name string) error {
	cmd := newCommand(player.commandAddPlaylist, name)
	res := player.commandDispatcher(cmd)

	if player.stateListener != nil {
		player.stateListener.PlaylistsChanged()
	}

	return res.err
}

// Add adds track or directory pointed by the path parameter to the playlist.
// If path represents a directory entry all containing items will be added recursive.
// This function expects playlist to be locked and synchronized by a caller.
func (player *Player) Add(name string, path *vfs.Path) error {
	cmd := newCommand(player.commandAdd, name, path)
	res := player.commandDispatcher(cmd)

	if player.stateListener != nil {
		player.stateListener.PlaylistChanged(name)
	}

	return res.err
}

// DeletePlaylist removes playlist by name.
func (player *Player) DeletePlaylist(name string) error {
	cmd := newCommand(player.commandDeletePlaylist, name)
	res := player.commandDispatcher(cmd)

	if player.stateListener != nil {
		player.stateListener.PlaylistsChanged()
	}

	return res.err
}

// Stop playing.
func (player *Player) Stop() {
	cmd := newCommand(player.commandStop, nil)
	player.commandDispatcher(cmd)
}

// Pause or resume player.
func (player *Player) Pause() {
	cmd := newCommand(player.commandPause, nil)
	player.commandDispatcher(cmd)
}

// Returns true if player is in paused state now.
func (player *Player) Paused() bool {
	return player.playingThread.Paused()
}

// commandDispatcher is a wrapper for syncronous communication with
// command routine.
func (player *Player) commandDispatcher(cmd *command) *result {
	player.commandChan <- cmd
	res := <-cmd.resultChan

	return res
}

// commandRoutine is the player commands handler goroutine.
// It doesn't hanlde any actual audio playing, and exists only
// to make player's public API methods syncronous. In this case
// player becomes threadsafe and all it's public methods can be invoked
// from different routines without any synchronization (commandChan
// do this job instead).
func (player *Player) commandRoutine() {
	for {
		cmd := <-player.commandChan
		cmd.resultChan <- player.dispatchCommand(cmd)
	}
}

// Call appropriate handler for the given command.
func (player *Player) dispatchCommand(cmd *command) *result {
	return cmd.method(cmd.args...)
}

// Start playing track in *vfs* playlist.
func (player *Player) commandPlayTrack(args ...interface{}) *result {
	track := args[0].(*vfs.Track)

	plist, _ := player.playlistByName(PlaylistVfs)
	plist.Clear()

	entries, err := track.Path.Parent().List()
	if err != nil {
		return newEmptyResult()
	}

	for _, entry := range entries {
		if track, ok := entry.(*vfs.Track); ok {
			plist.Append(track)
		}
	}

	// TODO: Find correct index. How? Compare Path + Part + Start?
	return player.commandPlayPlaylist(PlaylistVfs, 0)
}

// Start playing existing playlist from the given position.
func (player *Player) commandPlayPlaylist(args ...interface{}) *result {
	name := args[0].(string)
	track := args[1].(int)
	plist, err := player.playlistByName(name)
	if err != nil {
		return newErrorResult(err)
	}

	player.playingThread = playingthread.New(plist)
	err = player.playingThread.Run()
	if err != nil {
		return newErrorResult(err)
	}
	err = player.playingThread.Play(track)
	if err != nil {
		return newErrorResult(err)
	}

	return newEmptyResult()
}

// Returns playlists list.
func (player *Player) commandPlaylists(args ...interface{}) *result {
	// TODO: playlists should be copied.
	return newResult(player.playlists)
}

// Returns playlist content.
func (player *Player) commandPlaylist(args ...interface{}) *result {
	name := args[0].(string)

	plist, err := player.playlistByName(name)
	if err != nil {
		return newErrorResult(err)
	}

	if plist == player.currentPlaylist {
		plist.Lock()
		defer plist.Unlock()
	}

	// TODO: Tracks list should be copied.
	return newResult(plist)
}

// Creates new empty playlist with give name. Playlist name should be unique,
// so if playlist with given name exists error will be returned.
func (player *Player) commandAddPlaylist(args ...interface{}) *result {
	name := args[0].(string)

	if _, err := player.playlistByName(name); err == nil {
		return newErrorResult(fmt.Errorf("Playlist %s already exists.", name))
	}

	plist := playlist.New(name)
	if plist.System() {
		return newErrorResult(errors.New("System playlist can't be created."))
	}

	player.playlists = append(player.playlists, plist)

	return newEmptyResult()
}

// commandAdd adds object pointed by path to the end of the playlist.
func (player *Player) commandAdd(args ...interface{}) *result {
	name := args[0].(string)
	path := args[1].(*vfs.Path)

	plist, err := player.playlistByName(name)
	if err != nil {
		return newErrorResult(err)
	}

	if plist == player.currentPlaylist {
		plist.Lock()
		defer plist.Unlock()
	}

	player.add(plist, path)

	return newEmptyResult()
}

// Deletes existing playlist by name.
func (player *Player) commandDeletePlaylist(args ...interface{}) *result {
	// TODO: Stop playing if playing current playlist.
	// TODO: Clear current playlist on stop.
	name := args[0].(string)

	for i, playlist := range player.playlists {
		if playlist.Name() == name {
			if playlist.System() {
				return newErrorResult(fmt.Errorf("System playlist can't be deleted"))
			}

			player.playlists = append(player.playlists[:i],
				player.playlists[i+1:]...)
			break
		}
	}

	return newEmptyResult()
}

// Stop current playing player command.
func (player *Player) commandStop(args ...interface{}) *result {
	player.playingThread.Stop()

	return newEmptyResult()
}

// Pauses/Resumes playing process.
func (player *Player) commandPause(args ...interface{}) *result {
	player.playingThread.Pause()

	return newEmptyResult()
}

// add adds recursive path to the end of the playlist.
func (player *Player) add(plist *playlist.Playlist, path *vfs.Path) {
	isDir, err := path.IsDirectory()
	if err != nil {
		// TODO: Log error.
		// Ignore broken folders.
		return
	}

	if isDir {
		entries, err := path.List()
		if err != nil {
			// TODO: Log the error.
			// Ignore broken folders.
			return
		} else {
			for _, e := range entries {
				switch e.(type) {
				case *vfs.Track:
					plist.Append(e.(*vfs.Track))
				case *vfs.Directory:
					player.add(plist, e.(*vfs.Directory).Path)
				}
			}
		}
	} else {
		track, err := path.Track()
		if err != nil {
			// TODO: Log it.
		} else {
			plist.Append(track)
		}
	}
}

// playlistByName returns playlist for given name
// or nil if there is no such playlist exists.
func (player *Player) playlistByName(name string) (pl *playlist.Playlist, err error) {
	for _, plist := range player.playlists {
		if plist.Name() == name {
			return plist, nil
		}
	}

	return nil, fmt.Errorf("Playlist %s does not exist.", name)
}
