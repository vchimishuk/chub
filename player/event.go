// Copyright 2023 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package player

import (
	"github.com/vchimishuk/chub/serialize"
	"github.com/vchimishuk/chub/vfs"
)

type Event interface {
	Name() string
	Serialize() []serialize.Serializable
}

type StatusEvent struct {
	State    State
	Volume   int
	Plist    *Playlist
	PlistPos int
	Track    *vfs.Track
	TrackPos int
}

func (e *StatusEvent) Name() string {
	return "status"
}

func (e *StatusEvent) Serialize() []serialize.Serializable {
	st := map[string]any{}
	st["state"] = e.State.String()
	st["volume"] = e.Volume
	if e.State != StateStopped {
		st["playlist-duration"] = e.Plist.Duration()
		st["playlist-length"] = e.Plist.Len()
		st["playlist-name"] = e.Plist.Name()
		st["playlist-position"] = e.PlistPos
		st["track-album"] = e.Track.Tag.Album
		st["track-artist"] = e.Track.Tag.Artist
		st["track-length"] = e.Track.Length
		st["track-number"] = e.Track.Tag.Number
		st["track-path"] = e.Track.Path.String()
		st["track-position"] = e.TrackPos
		st["track-title"] = e.Track.Tag.Title
		st["track-year"] = e.Track.Tag.Year
	}

	return []serialize.Serializable{serialize.Wrap(st)}
}

type PlistCreateEvent struct {
	Plist string
}

func (e *PlistCreateEvent) Name() string {
	return "create-playlist"
}

func (e *PlistCreateEvent) Serialize() []serialize.Serializable {
	return []serialize.Serializable{serialize.Wrap(map[string]any{
		"name": e.Plist,
	})}
}

type PlistDeleteEvent struct {
	Plist string
}

func (e *PlistDeleteEvent) Name() string {
	return "delete-playlist"
}

func (e *PlistDeleteEvent) Serialize() []serialize.Serializable {
	return []serialize.Serializable{serialize.Wrap(map[string]any{
		"name": e.Plist,
	})}
}

type PlistRenameEvent struct {
	From string
	To   string
}

func (e *PlistRenameEvent) Name() string {
	return "playlist-rename"
}

func (e *PlistRenameEvent) Serialize() []serialize.Serializable {
	return []serialize.Serializable{serialize.Wrap(map[string]any{
		"from": e.From,
		"to":   e.To,
	})}
}
