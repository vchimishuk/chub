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

package player

import (
	"errors"
	"sync"

	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/vfs"
)

const (
	vfsPlistName = "*vfs*"
)

type Event string

const (
	EventStatus Event = "status"
	EventVolume Event = "volume"
)

type EventHandler func(e Event, args []interface{})

type Player struct {
	// Mutex guards plists and curPlist fields.
	// Any manipulation on that fields must be guarded with this mutex.
	plistsMu sync.RWMutex
	plists   map[string]*Playlist
	curPlist *Playlist
	// Used output driver.
	output Output
	// Playing thread, which manages decode-output loop.
	pt           *playingThread
	eventHandler EventHandler
}

func New(fmts []format.Format, output Output) *Player {
	p := &Player{
		plists:   make(map[string]*Playlist),
		curPlist: NewPlaylist(vfsPlistName),
		output:   output,
		pt:       newPlayingThread(fmts, output),
	}
	p.pt.SetStatusHandler(func(s *Status) {
		p.notify(EventStatus, s)
	})
	p.pt.Start()
	p.pt.SetPlaylist(p.curPlist)

	return p
}

func (p *Player) SetEventHandler(h EventHandler) {
	p.eventHandler = h
}

func (p *Player) Close() {
	p.pt.Close()
}

func (p *Player) Play(path *vfs.Path) error {
	// Fill VFS playlist.
	var dir *vfs.Path
	if path.IsDir() {
		dir = path
	} else {
		d, err := path.Parent()
		if err != nil {
			return err
		}
		dir = d
	}
	entries, err := dir.List()
	if err != nil {
		return err
	}

	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	p.curPlist = NewPlaylist(vfsPlistName)
	pos := 0
	i := 0
	for _, e := range entries {
		if !e.IsDir() {
			t := e.(*vfs.Track)
			// TODO: Avoid coping plist all the time,
			//       use bulk append instead.
			p.curPlist = p.curPlist.Append(t)
			if *path == *t.Path {
				pos = i
			}
			i++
		}
	}
	p.pt.Play(p.curPlist, pos)

	return nil
}

func (p *Player) Stop() {
	p.pt.Stop()
}

func (p *Player) Pause() {
	p.pt.Pause()
}

func (p *Player) Next() {
	p.pt.Next()
}

func (p *Player) Prev() {
	p.pt.Prev()
}

func (p *Player) Append(name string, path *vfs.Path) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.userPlist(name)
	if err != nil {
		return err
	}

	// TODO: Use walk style here to avoid extra array creation.
	tracks, err := listDirRec(path)
	if err != nil {
		return err
	}
	p.replace(name, pl.Append(tracks...))

	// TODO: Add new tracks parameter to notification,
	//       so client can update his playlist version
	//       without requesting new version.
	//       The same for other similar functions.
	// go p.notify(PlaylistEvent, plist)

	return nil
}

func (p *Player) Clear(name string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.userPlist(name)
	if err != nil {
		return err
	}
	p.replace(name, pl.Clear())

	return nil
}

func (p *Player) Create(name string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	_, err := p.userPlist(name)
	if err == nil {
		return errors.New("already exists")
	}

	p.plists[name] = NewPlaylist(name)

	// go p.notify(PlaylistsEvent, name)

	return nil
}

func (p *Player) Delete(name string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.userPlist(name)
	if err != nil {
		return err
	}

	delete(p.plists, name)
	if pl.Name() == p.curPlist.Name() {
		p.pt.Stop()
		p.curPlist = nil
	}

	// TODO: go p.notify(PlaylistsEvent, p.Playlists())

	return nil
}

func (p *Player) Playlist(name string) (*Playlist, error) {
	p.plistsMu.RLock()
	defer p.plistsMu.RUnlock()

	pl, err := p.userPlist(name)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *Player) Rename(from string, to string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.userPlist(from)
	if err != nil {
		return err
	}

	p.replace(from, pl.SetName(to))

	// go p.notify(PlaylistsEvent, p.Playlists())

	return nil
}

func (p *Player) Playlists() []*Playlist {
	p.plistsMu.RLock()
	defer p.plistsMu.RUnlock()

	plists := make([]*Playlist, 0, len(p.plists))
	for _, pl := range p.plists {
		plists = append(plists, pl)
	}

	return plists
}

func (p *Player) Status() *Status {
	return p.pt.Status()
}

func (p *Player) userPlist(name string) (*Playlist, error) {
	if name == vfsPlistName {
		return nil, errors.New("invalid playlist")
	}

	pl, ok := p.plists[name]
	if !ok {
		return nil, errors.New("invalid playlist")
	}

	return pl, nil
}

func (p *Player) replace(name string, pl *Playlist) {
	delete(p.plists, name)

	p.plists[pl.Name()] = pl
	if p.curPlist.Name() == name {
		p.curPlist = pl
		p.pt.SetPlaylist(pl)
	}
}

func (p *Player) notify(e Event, args ...interface{}) {
	if p.eventHandler != nil {
		go p.eventHandler(e, args)
	}
}

func listDirRec(path *vfs.Path) ([]*vfs.Track, error) {
	var tracks []*vfs.Track

	if path.IsDir() {
		entries, err := path.List()
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			if dir, ok := e.(*vfs.Dir); ok {
				t, err := listDirRec(dir.Path)
				if err != nil {
					return nil, err
				}
				tracks = append(tracks, t...)
			} else if track, ok := e.(*vfs.Track); ok {
				tracks = append(tracks, track)
			} else {
				panic(nil)
			}
		}
	} else {
		track, err := path.Track()
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}
