// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package cmd

import (
	"errors"
	"fmt"
)

const (
	// Create new playlist.
	cmdCreatePlaylist = "create-playlist"
	// Delete existing playlist.
	cmdDeletePlaylist = "delete-playlist"
	// Stop the server.
	cmdKill = "kill"
	// Show directory contents.
	cmdList = "list"
	// Play next track in the current playing playlist.
	cmdNext = "next"
	// Toggle paused state.
	cmdPause = "pause"
	// Do nothing, just returns "OK" response.
	cmdPing = "ping"
	// Play a path (track or directory) from VFS.
	cmdPlay = "play"
	// Add new track or folder to the end of the playlist.
	cmdPlaylistAppend = "playlist-append"
	// Remove all items from playlist.
	cmdPlaylistClear = "playlist-clear"
	// Delete items from playlist.
	cmdPlaylistDelete = "playlist-delete"
	// Show playlist tracks.
	cmdPlaylistList = "playlist-list"
	// Start playing given playlist.
	cmdPlaylistPlay = "playlist-play"
	// Move items inside playlist.
	cmdPlaylistRemove = "playlist-move"
	// Rename playlist.
	cmdPlaylistRename = "rename-playlist"
	// Show existing playlists list.
	cmdPlaylists = "playlists"
	// Play previous track in the current playling playlist.
	cmdPrev = "prev"
	// Disconnect from server.
	cmdQuit = "quit"
	// Set/toggle repeat mode.
	cmdRepeat = "repeat"
	// Returns player's current state (playback status, volume, etc.).
	cmdStatus = "status"
	// Seek current playing track time to specified time offset.
	cmdSeek = "seek"
	// Stop playing if active.
	cmdStop = "stop"
	// Change volume level.
	cmdVolumn = "volume"
)

type command struct {
	name string
	args []interface{}
}

func parseCommand(str string) (*command, error) {
	args := []interface{}{}
	var err error
	s := newScanner(str)

	if !s.HasNext() {
		return nil, errors.New("invalid command")
	}
	name, err := s.NextString()
	if err != nil {
		return nil, errors.New("invalid command")
	}

	switch name {
	// One int argument command.
	case cmdSeek:
		t, e := s.NextInt()
		r := false
		if s.HasNext() {
			r, e = s.NextBool()
		} else {
			if t < 0 {
				e = errors.New("negative time")
			}
		}
		args = []interface{}{t, r}
		err = e
	// One string argument command.
	case cmdCreatePlaylist, cmdList, cmdPlay, cmdPlaylistClear:
		fallthrough
	case cmdPlaylistDelete, cmdPlaylistList:
		path, e := s.NextString()
		args = []interface{}{path}
		err = e
	// Two string arguments command.
	case cmdPlaylistAppend, cmdPlaylistRename:
		path := ""
		name, e := s.NextString()
		if e == nil {
			path, e = s.NextString()
		}
		args = []interface{}{name, path}
		err = e
	// Argumentless command.
	case cmdKill, cmdNext, cmdPause, cmdPing, cmdPlaylists:
		fallthrough
	case cmdPrev, cmdQuit, cmdStatus, cmdStop:
		// Argumentless command.
	default:
		return nil, errors.New("unsupported command")
	}

	if s.HasNext() {
		err = errors.New("to many arguments")
	}
	if err != nil {
		return nil, fmt.Errorf("invalid command format")
	}

	return &command{name: name, args: args}, nil
}
