// player package is the core of the program: it manages playlists and player's state.
package player

import (
	"../audio"
	"../vfs"
	"./playlist"
	"errors"
	"fmt"
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

// Playing process communication command response.
type response struct {
	// Error.
	err error
	// Command processing result itself.
	arguments interface{}
}

// Playing process communication command.
type command struct {
	code         int
	arguments    []interface{}
	responseChan chan *response
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
	// Channel for communication with playingRoutine.
	playingChan chan int
	// Playing process current status.
	state int
	// Output driver.
	output audio.Output
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
	_, err := player.command(commandPlayTrack, track)

	return err
}

// PlayPlaylist starts playing given playlist from the given position in the playlist.
// First track in the playlist has 0 position.
func (player *Player) PlayPlaylist(name string, track int) error {
	_, err := player.command(commandPlayPlaylist)

	return err
}

// Playlists eturns list with all playlists in the system.
func (player *Player) Playlists() []*playlist.Playlist {
	res, _ := player.command(commandPlaylists)

	return res.([]*playlist.Playlist)
}

// AddPlaylist creates new playlist.
func (player *Player) AddPlaylist(name string) error {
	_, err := player.command(commandAddPlaylist, name)

	return err
}

// AppendTrack appends given track to the playlist.
func (player *Player) AppendTrack(name string, track *vfs.Track) error {
	_, err := player.command(commandAppendTrack, name, track)

	return err
}

// DeletePlaylist removes playlist by name.
func (player *Player) DeletePlaylist(name string) error {
	_, err := player.command(commandDeletePlaylist, name)

	return err
}

// Stop playing.
func (player *Player) Stop() {
	player.command(commandStop)
}

// Pause or resume player.
func (player *Player) Pause() {
	player.command(commandPause)
}

// Command allows to communicate with Player by sending him commands.
func (player *Player) command(cmd int, args ...interface{}) (res interface{}, err error) {
	c := &command{code: cmd, arguments: args, responseChan: make(chan *response)}

	player.commandChan <- c
	resp := <-c.responseChan

	return resp.arguments, resp.err
}

// Start playing procedure for the given playlist.
func (player *Player) playPlaylist(plist *playlist.Playlist, startPos int) {
	go player.playingRoutine(plist, startPos)
}

// playingProcess is the player commands handler.
func (player *Player) commandRoutine() {
	for {
		cmd := <-player.commandChan
		cmd.responseChan <- player.dispatchCommand(cmd)
	}
}

// playingRoutine does data decoding and sound output.
// This routine starts playing given playlist from the startPos.
func (player *Player) playingRoutine(plist *playlist.Playlist, startPos int) {
	// TODO: This function is not correct, -- playlist is not protected well.

	plist.Lock()
	defer plist.Unlock()

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

// Call appropriate handler for the given command. 
func (player *Player) dispatchCommand(cmd *command) *response {
	r := new(response)

	// TODO: Replace this switch with some kind of map solution.
	switch cmd.code {
	case commandPlayTrack:
		r.arguments, r.err = player.commandPlayTrack(
			cmd.arguments[0].(*vfs.Track))
	case commandPlayPlaylist:
		r.arguments, r.err = player.commandPlayPlaylist(
			cmd.arguments[0].(string),
			cmd.arguments[1].(int))
	case commandPlaylists:
		r.arguments, r.err = player.commandPlaylistsList()
	case commandAddPlaylist:
		r.arguments, r.err = player.commandPlaylistAdd(
			cmd.arguments[0].(string))
	case commandAppendTrack:
		r.arguments, r.err = player.commandPlaylistAppendTrack(
			cmd.arguments[0].(string),
			cmd.arguments[1].(*vfs.Track))
	case commandDeletePlaylist:
		r.arguments, r.err = player.commandPlaylistDelete(
			cmd.arguments[0].(string))
	case commandStop:
		r.arguments, r.err = player.commandStop()
	case commandPause:
		r.arguments, r.err = player.commandPause()

	// case CMD_PLAYLIST_ADD:
	//	r.err = player.cmdPlaylistAdd()
	default:
		r.err = fmt.Errorf("Unsupported command %s.", cmd.code)
	}

	return r
}

// Start playing track in *vfs* playlist.
func (player *Player) commandPlayTrack(track *vfs.Track) (_ interface{}, err error) {
	plist, _ := player.playlistByName(PlaylistVfs)
	plist.Clear()

	entries, err := vfs.NewForPath(track.Path.Parent()).List()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if track, ok := entry.(*vfs.Track); ok {
			plist.Append(track)
		}
	}

	// TODO: Find correct index. How? Compare Path + Part + Start?
	player.playPlaylist(plist, 0)

	return nil, nil
}

// Start playing existing playlist from the given position.
func (player *Player) commandPlayPlaylist(name string, track int) (_ interface{}, err error) {
	// TODO:

	return nil, nil
}

// Returns playlists list.
func (player *Player) commandPlaylistsList() (data []*playlist.Playlist, err error) {
	return player.playlists, nil
}

// Creates new empty playlist with give name. Playlist name should be unique,
// so if playlist with given name exists error will be returned. 
func (player *Player) commandPlaylistAdd(name string) (_ interface{}, err error) {
	if _, err := player.playlistByName(name); err == nil {
		return nil, fmt.Errorf("Playlist %s already exists.", name)
	}

	plist := playlist.New(name)
	if plist.System() {
		return nil, errors.New("System playlist can't be created.")
	}

	player.playlists = append(player.playlists, plist)

	return nil, nil
}

// Append one track to the playlist.
func (player *Player) commandPlaylistAppendTrack(name string, track *vfs.Track) (_ interface{}, err error) {
	plist, err := player.playlistByName(name)
	if err != nil {
		return nil, err
	}

	plist.Append(track)

	return nil, nil
}

// Deletes existing playlist by name.
func (player *Player) commandPlaylistDelete(name string) (_ interface{}, err error) {
	// TODO: Stop playing if playing current playlist.
	// TODO: Clear current playlist on stop.

	for i, playlist := range player.playlists {
		if playlist.Name() == name {
			if playlist.System() {
				return nil, fmt.Errorf("System playlist can't be deleted")
			}

			player.playlists = append(player.playlists[:i],
				player.playlists[i+1:]...)
			break
		}
	}

	return nil, nil
}

// Stop current playing player command.
func (player *Player) commandStop() (_ interface{}, _ error) {
	if player.state != stateStopped {
		player.playingChan <- playingCommandStop
	}

	return nil, nil
}

// Pauses/Resumes playing process.
func (player *Player) commandPause() (_ interface{}, _ error) {
	if player.state != stateStopped {
		player.playingChan <- playingCommandPause
	}

	return nil, nil
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
