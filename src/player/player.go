// player package is the core of the program: it manages playlists and player's state.
package player

import (
	"../audio"
	"../vfs"
	"./playlist"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	PlaylistVfs = "*vfs*"
)

const (
	stateStopped int = iota
	statePlaying
	statePaused
)

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
	// Channel for communication with playingRoutine.
	playingChan chan int
	// Playing process current status.
	state int
	// Output driver.
	output audio.Output
	// Playlist modification lock. Every time playlist when you need to make
	// any playlist modification you should get this lock before. Also lock
	// should be gotten when you select next song from the playlist,
	// e.g. playing goroutine have to lock the current playlist during
	// moving to the next song in the current playlist.
	playlistLock sync.Mutex
}

// New returns a newly created Player object.
func New() *Player {
	p := new(Player)
	p.commandChan = make(chan *command, 10)
	p.playingChan = make(chan int)
	// Create some predefined system playlists.
	p.playlists = append(p.playlists, playlist.New(PlaylistVfs))

	return p
}

// Run starts Player execution.
func (player *Player) Run() (err error) {
	player.output, err = audio.GetOutput()
	if err != nil {
		return err
	}

	go player.commandRoutine()

	return nil
}

// PlayTrack starts playing track in *vfs* playlist.
func (player *Player) PlayTrack(track *vfs.Track) error {
	cmd := newCommand(player.commandPlayTrack, []interface{}{track})
	res := player.commandDispatcher(cmd)

	return res.err
}

// PlayPlaylist starts playing given playlist from the given position in the playlist.
// First track in the playlist has 0 position.
func (player *Player) PlayPlaylist(name string, track int) error {
	cmd := newCommand(player.commandPlayPlaylist,
		[]interface{}{name, track})
	res := player.commandDispatcher(cmd)

	return res.err
}

// Playlists eturns list with all playlists in the system.
func (player *Player) Playlists() []*playlist.Playlist {
	cmd := newCommand(player.commandPlaylists,
		[]interface{}{})
	res := player.commandDispatcher(cmd)
	plists := res.args[0].([]*playlist.Playlist)

	return plists
}

// AddPlaylist creates new playlist.
func (player *Player) AddPlaylist(name string) error {
	cmd := newCommand(player.commandAddPlaylist,
		[]interface{}{name})
	res := player.commandDispatcher(cmd)

	return res.err
}

// AppendTrack appends given track to the playlist.
func (player *Player) AppendTrack(name string, track *vfs.Track) error {
	cmd := newCommand(player.commandAppendTrack,
		[]interface{}{name, track})
	res := player.commandDispatcher(cmd)

	return res.err
}

// DeletePlaylist removes playlist by name.
func (player *Player) DeletePlaylist(name string) error {
	cmd := newCommand(player.commandDeletePlaylist,
		[]interface{}{name})
	res := player.commandDispatcher(cmd)

	return res.err
}

// Stop playing.
func (player *Player) Stop() {
	cmd := newCommand(player.commandStop,
		[]interface{}{})
	player.commandDispatcher(cmd)
}

// Pause or resume player.
func (player *Player) Pause() {
	cmd := newCommand(player.commandPause,
		[]interface{}{})
	player.commandDispatcher(cmd)
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
	return cmd.method(cmd.args)
}

// Start playing procedure for the given playlist.
func (player *Player) playPlaylist(plist *playlist.Playlist, startPos int) {
	go player.playingRoutine(plist, startPos)
}

// playingRoutine does data decoding and sound output.
// This routine starts playing given playlist from the startPos.
func (player *Player) playingRoutine(plist *playlist.Playlist, startPos int) {
	// TODO: This function is not correct, -- playlist is not protected well.

	// TODO: plist.Lock()
	// TODO: defer plist.Unlock()

	if plist.Tracks().Len() == 0 || plist.Tracks().Len() <= startPos {
		return
	}

	err := player.output.Open()
	if err != nil {
		// TODO: Log this error.
		return
	}
	defer player.output.Close()
	player.output.SetSampleRate(44100)
	player.output.SetChannels(2)

	// Find correct entry for start.
	entry := plist.Tracks().Front()
	for i := 0; i < startPos; i++ {
		entry = entry.Next()
	}

	player.currentPlaylist = plist
	playNext := true

	for playNext && entry != nil {
		player.currentTrack = entry.Value.(*vfs.Track)
		decoder, err := audio.GetDecoder(player.currentTrack)

		if err == nil {
			err = decoder.Open(player.currentTrack) // XXX:
			if err != nil {
				// TODO: Just log it and continue.
				panic("decoder.Open() failed.")
			}
			playNext = player.doPlay(decoder, player.output)
		} else {
			// Just skip this track on any problem.
			// TODO: Log it.
		}

		entry = entry.Next()
	}
}

// Do actual decode -> output process.
// Returns false if playing process should be stopped (i.e. stop command was received).
func (player *Player) doPlay(decoder audio.Decoder, output audio.Output) bool {
	// TODO: Refactor all this crappy function.
	bufAvailable := make(chan bool)
	bufAvailableChecker := func() {
		// TODO: Add some mutex to be sure that only one instance of the
		//       groutine is running.
		for player.state != stateStopped {
			ready, err := output.Wait(100)

			if err != nil {
				// Sometimes Wait failed, so just wait some time and retry.
				time.Sleep(100 * time.Millisecond)
			} else if ready && player.state == statePlaying {
				bufAvailable <- true
			}
		}
	}

	player.state = statePlaying

	// TODO: Check output HW params and maybe apply new ones.

	go bufAvailableChecker()

	defer fmt.Printf("exiting doPlay()\n")

	for {
		select {
		case cmd := <-player.playingChan:
			switch cmd {
			case playingCommandStop:
				player.state = stateStopped
			case playingCommandPause:
				switch player.state {
				case statePlaying:
					player.state = statePaused
				case statePaused:
					player.state = statePlaying
				default:
					panic("Illegal player state found.")
				}
			}
		case <-bufAvailable:
			// Do nothing, just wake up.
		}

		switch player.state {
		case statePlaying:
			size, _ := output.AvailUpdate()

			buf := make([]byte, size)
			// TODO: Error processing.
			// TODO: When track is finished?
			read, _ := decoder.Read(buf)
			// written, err := output.Write(buf[:read])
			output.Write(buf[:read])

			// fmt.Printf("size: %d\n", size)
			// fmt.Printf("read: %d\n", read)
			// fmt.Printf("written: %d\n", written)
			// fmt.Printf("err: %v\n", err)
		case stateStopped:
			return false
		case statePaused:
			// Do nothing here, just sleep with select till
			// playingCommandPause wakes us up.
		}
	}

	return true
}

// Start playing track in *vfs* playlist.
func (player *Player) commandPlayTrack(args []interface{}) *result {
	track := args[0].(*vfs.Track)

	plist, _ := player.playlistByName(PlaylistVfs)
	plist.Clear()

	entries, err := vfs.NewForPath(track.Path.Parent()).List()
	if err != nil {
		return newEmptyResult()
	}

	for _, entry := range entries {
		if track, ok := entry.(*vfs.Track); ok {
			plist.Append(track)
		}
	}

	// TODO: Find correct index. How? Compare Path + Part + Start?
	player.playPlaylist(plist, 0)

	return newEmptyResult()
}

// Start playing existing playlist from the given position.
func (player *Player) commandPlayPlaylist(args []interface{}) *result {
	// plistName := args[0].(string)
	// track := args[1].(int)

	// TODO:

	return newEmptyResult()
}

// Returns playlists list.
func (player *Player) commandPlaylists(args []interface{}) *result {
	return newResult(player.playlists)
}

// Creates new empty playlist with give name. Playlist name should be unique,
// so if playlist with given name exists error will be returned.
func (player *Player) commandAddPlaylist(args []interface{}) *result {
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

// Append one track to the playlist.
func (player *Player) commandAppendTrack(args []interface{}) *result {
	name := args[0].(string)
	track := args[1].(*vfs.Track)

	plist, err := player.playlistByName(name)
	if err != nil {
		return newErrorResult(err)
	}

	player.playlistLock.Lock()
	defer player.playlistLock.Unlock()
	plist.Append(track)

	return newEmptyResult()
}

// Deletes existing playlist by name.
func (player *Player) commandDeletePlaylist(args []interface{}) *result {
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
func (player *Player) commandStop(args []interface{}) *result {
	if player.state != stateStopped {
		player.playingChan <- playingCommandStop
	}

	return newEmptyResult()
}

// Pauses/Resumes playing process.
func (player *Player) commandPause(args []interface{}) *result {
	if player.state != stateStopped {
		player.playingChan <- playingCommandPause
	}

	return newEmptyResult()
}

// playlistByName returns playlist for given name
// or nil if there is no such playlist exists.
func (player *Player) playlistByName(name string) (pl *playlist.Playlist, err error) {
	for _, plist := range player.playlists {
		if plist.Name() == name {
			return plist, nil
		}
	}

	return nil, fmt.Errorf("Playlist %s not found.", name)
}
