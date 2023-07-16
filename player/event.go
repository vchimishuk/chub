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

import "github.com/vchimishuk/chub/serialize"

type Event interface {
	Name() string
	Body() []serialize.Serializable
}

type PlistCreateEvent struct {
	Plist string
}

func (e *PlistCreateEvent) Name() string {
	return "create-playlist"
}

func (e *PlistCreateEvent) Body() []serialize.Serializable {
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

func (e *PlistDeleteEvent) Body() []serialize.Serializable {
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

func (e *PlistRenameEvent) Body() []serialize.Serializable {
	return []serialize.Serializable{serialize.Wrap(map[string]any{
		"from": e.From,
		"to":   e.To,
	})}
}
