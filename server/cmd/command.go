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
// along with Chub.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"errors"
	"fmt"
)

const (
	// TODO: Replace BACKWARD & FORWARD with SEEK command.

	// Seek playing track position backward.
	cmdBackward = "BACKWARD"
	// Seek playing track position forward.
	cmdForward = "FORWARD"
	// Stop the server.
	cmdKill = "KILL"
	// Show directory contents.
	cmdLs = "LS"
	// Play next track in the current playing playlist.
	cmdNext = "NEXT"
	// Toggle paused state.
	cmdPause = "PAUSE"
	// Do nothing, just returns "OK" response.
	cmdPing = "PING"
	// Play a path (track or directory) from VFS.
	cmdPlay = "PLAY"
	// Add new tracks or folder to the end of the playlist.
	cmdPlaylistAppend = "PLAYLIST_APPEND"
	// Remove all items from playlist.
	cmdPlaylistClear = "PLAYLIST_CLEAR"
	// Create new playlist.
	cmdPlaylistCreate = "PLAYLIST_CREATE"
	// Delete existing playlist.
	cmdPlaylistDelete = "PLAYLIST_DELETE"
	// Show playlist tracks.
	cmdPlaylistList = "PLAYLIST_LIST"
	// Start playing given playlist.
	cmdPlaylistPlay = "PLAYLIST_PLAY"
	// Remove items from playlist.
	cmdPlaylistRemove = "PLAYLIST_REMOVE"
	// Rename playlist.
	cmdPlaylistRename = "PLAYLIST_RENAME"
	// Show existing playlists.
	cmdPlaylistsList = "PLAYLISTS_LIST"
	// Play previous track in the current playling playlist.
	cmdPrev = "PREV"
	// Disconnect from server.
	cmdQuit = "QUIT"
	// Set/toggle repeat mode.
	cmdRepeat = "REPEAT"
	// Returns player's current state (playback status, volume, etc.).
	cmdState = "STATE"
	// Change volume level.
	cmdVolumn = "VOLUME"
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
	case cmdLs, cmdPlay, cmdPlaylistClear, cmdPlaylistCreate:
		fallthrough
	case cmdPlaylistDelete, cmdPlaylistList:
		// One string argument command.
		path, e := s.NextString()
		args = []interface{}{path}
		err = e
	case cmdPlaylistAppend, cmdPlaylistRename:
		// Two string arguments command.
		path := ""
		name, e := s.NextString()
		if e == nil {
			path, e = s.NextString()
		}
		args = []interface{}{name, path}
		err = e
	case cmdKill, cmdPing, cmdPlaylistsList, cmdQuit:
		// Argumentless command.
	default:
		return nil, errors.New("unsupported command")
	}

	if s.HasNext() {
		err = errors.New("to many arguments")
	}
	if err != nil {
		fmt.Printf(":::%V\n", err)
		return nil, fmt.Errorf("invalid command format")
	}

	return &command{name: name, args: args}, nil
}
