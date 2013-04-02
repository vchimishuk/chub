// player package is the core of the program: it manages playlists and player's state.
package player

import (
	"../vfs"
	"./playlist"
	"errors"
	"fmt"
)

const (
	PlaylistVfs = "*vfs*"
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
	// Channel to communicate player with. Client code can
	// write commands and read responses to/from the channel.
	playingChan chan *command
}

// New returns a newly created Player object.
func New() *Player {
	p := new(Player)
	p.playingChan = make(chan *command, 10)
	// Create some predefined system playlists.
	p.playlists = append(p.playlists, playlist.New(PlaylistVfs))

	return p
}

// Run starts Player execution.
func (player *Player) Run() {
	go player.playingProcess()
}

// Command allows to communicate with Player by sending him commands.
func (player *Player) Command(cmd int, args ...interface{}) (res interface{}, err error) {
	c := &command{code: cmd, arguments: args, responseChan: make(chan *response, 1)}

	player.playingChan <- c
	resp := <-c.responseChan

	return resp.arguments, resp.err
}

// playingProcess is the core of the player. It runs in goroutine
// and does the playing intself. Outer world can affects to playing process by
// sending commands via playingChan of the Player struct.
func (player *Player) playingProcess() {
	for {
		cmd := <-player.playingChan
		r := new(response)

		// TODO: Replace this switch with some kind of map solution.
		switch cmd.code {
		case CommandPlayTrack:
			r.arguments, r.err = player.commandPlayTrack(
				cmd.arguments[0].(*vfs.Track))
		case CommandPlaylistsList:
			r.arguments, r.err = player.commandPlaylistsList()
		case CommandPlaylistAdd:
			r.arguments, r.err = player.commandPlaylistAdd(
				cmd.arguments[0].(string))
		case CommandPlaylistAppendTrack:
			r.arguments, r.err = player.commandPlaylistAppendTrack(
				cmd.arguments[0].(string),
				cmd.arguments[1].(*vfs.Track))
		case CommandPlaylistDelete:
			r.arguments, r.err = player.commandPlaylistDelete(
				cmd.arguments[0].(string))

		// case CMD_PLAYLIST_ADD:
		//	r.err = player.cmdPlaylistAdd()
		default:
			r.err = fmt.Errorf("Unsupported command %s.", cmd.code)
		}

		cmd.responseChan <- r
	}
}

// Start playing track in *vfs* playlist.
func (player *Player) commandPlayTrack(track *vfs.Track) (_ interface{}, err error) {
	pl, _ := player.playlistByName(PlaylistVfs)
	pl.Clear()

	entries, err := vfs.NewForPath(track.Path.Parent()).List()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if track, ok := entry.(*vfs.Track); ok {
			pl.Append(track)
		}
	}

	// TODO: 1. Set current playlist (from which track is playing).
	//       2. And start playing.

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

	pl := playlist.New(name)
	if pl.System() {
		return nil, errors.New("System playlist can't be created.")
	}

	player.playlists = append(player.playlists, pl)

	return nil, nil
}

// Append one track to the playlist.
func (player *Player) commandPlaylistAppendTrack(name string, track *vfs.Track) (_ interface{}, err error) {
	pl, err := player.playlistByName(name)
	if err != nil {
		return nil, err
	}

	pl.Append(track)

	return nil, nil
}

// Deletes existing playlist by name.
func (player *Player) commandPlaylistDelete(name string) (_ interface{}, err error) {
	// TODO: Stop playing if playing current playlist.

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

// playlistByName returns playlist for given name
// or nil if there is no such playlist exists.
func (player *Player) playlistByName(name string) (pl *playlist.Playlist, err error) {
	for _, pl := range player.playlists {
		if pl.Name() == name {
			return pl, nil
		}
	}

	return nil, fmt.Errorf("Playlist %s not found.", name)
}

// --------------------

// Player mutex. All public player commands should be protected with this mutex lock.
// var mutex sync.Mutex

// thread is the main player thread (goroutine wrapper).
// var thread *playingThread

// // Playlist returns playlist object by name.
// func Playlist(name string) (playlist *playlist.Playlist, err os.Error) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	playlist, err = getPlaylistByName(name)

// 	return
// }

// // Play start playing existing a track form an existing playlist.
// func Play(playlistName string, trackNumber int) os.Error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	pl, err := getPlaylistByName(playlistName)
// 	if err != nil {
// 		return err
// 	}

// 	if trackNumber < 0 || trackNumber >= pl.Len() {
// 		return os.NewError(fmt.Sprintf("Playlist '%s' has no track number %d.",
// 			playlistName, trackNumber))
// 	}

// 	thread.Play(pl.Track(trackNumber))

// 	return nil
// }

// // Pause pause or unpause playing process.
// func Pause() {
// 	thread.Pause()
// }

// // Stop closes playing processes, frees resources.
// // This function should be called before exiting program.
// func Stop() {
// 	thread.Stop()
// }

// // Package init function.
// func init() {
// 	// Audio tagreaders.
// 	audio.RegisterTagReaderFactory(ogg.NewTagReader)
// 	// Audio outputs.
// 	audio.RegisterOutput(alsa.DriverName, alsa.New)
// 	// Audio decoders.
// 	audio.RegisterDecoderFactory(ogg.NewDecoder)

// 	// Playists
// 	playlists = make([]*playlist.Playlist, 0)
// 	// We have one system (predefined) playlist, -- *vfs*.
// 	playlists = append(playlists, playlist.New(vfs.PlaylistName))

// 	// Create and start playing thread.
// 	thread = newPlayingThread()
// 	thread.Start()
// }
