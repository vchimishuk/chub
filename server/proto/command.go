// Copyright 2016, 2023 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of Chub.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

package proto

const (
	// Create new playlist.
	CreatePlaylist = "create-playlist"
	// Delete existing playlist.
	DeletePlaylist = "delete-playlist"
	// Enable or disable events notification for the current connection.
	Events = "events"
	// Stop the server.
	Kill = "kill"
	// Show directory contents.
	List = "list"
	// Play next track in the current playing playlist.
	Next = "next"
	// Toggle paused state.
	Pause = "pause"
	// Do nothing, just returns "OK" response.
	Ping = "ping"
	// Play a path (track or directory) from VFS.
	Play = "play"
	// Add new track or folder to the end of the playlist.
	PlaylistAppend = "playlist-append"
	// Remove all items from playlist.
	PlaylistClear = "playlist-clear"
	// Delete items from playlist.
	PlaylistDelete = "playlist-delete"
	// Show playlist tracks.
	PlaylistList = "playlist-list"
	// Start playing given playlist.
	PlaylistPlay = "playlist-play"
	// Move items inside playlist.
	PlaylistRemove = "playlist-move"
	// Rename playlist.
	PlaylistRename = "rename-playlist"
	// Show existing playlists list.
	Playlists = "playlists"
	// Play previous track in the current playling playlist.
	Prev = "prev"
	// Disconnect from server.
	Quit = "quit"
	// Set/toggle repeat mode.
	Repeat = "repeat"
	// Returns player's current state (playback status, volume, etc.).
	Status = "status"
	// Seek current playing track time to specified time offset.
	Seek = "seek"
	// Stop playing if active.
	Stop = "stop"
	// Change volume level.
	Volumn = "volume"
)

type Command struct {
	Name string
	Args []interface{}
}

func ParseCommand(str string) (*Command, error) {
	args := []interface{}{}
	var err error
	s := newScanner(str)

	if !s.HasNext() {
		return nil, newError("invalid command")
	}
	name, err := s.NextString()
	if err != nil {
		return nil, newError("invalid command")
	}

	switch name {
	// TODO: Do not need boolean parameter for Seek.
	case Seek:
		t, e := s.NextInt()
		r := false
		if s.HasNext() {
			r, e = s.NextBool()
		} else {
			if t < 0 {
				e = newError("negative time")
			}
		}
		args = []interface{}{t, r}
		err = e
	// One bool argument command
	case Events:
		b, e := s.NextBool()
		args = []interface{}{b}
		err = e
	// One string argument command.
	case CreatePlaylist, DeletePlaylist, List, Play, PlaylistClear:
		fallthrough
	case PlaylistDelete, PlaylistList:
		p, e := s.NextString()
		args = []interface{}{p}
		err = e
	// Two string arguments command.
	case PlaylistAppend, PlaylistRename:
		b := ""
		a, e := s.NextString()
		if e == nil {
			b, e = s.NextString()
		}
		args = []interface{}{a, b}
		err = e
	// Argumentless command.
	case Kill, Next, Pause, Ping, Playlists:
		fallthrough
	case Prev, Quit, Status, Stop:
	default:
		return nil, newError("unsupported command")
	}

	if s.HasNext() {
		err = newError("to many arguments")
	}
	if err != nil {
		return nil, newError("invalid command format")
	}

	return &Command{Name: name, Args: args}, nil
}
