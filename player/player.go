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

	eventsChSize = 16
)

// TODO: Block active playlist editing.

type Player struct {
	// Mutex guards plists and curPlist fields.
	// Any manipulation on that fields must be guarded with this mutex.
	plistsMu sync.RWMutex
	plists   map[string]*Playlist
	curPlist *Playlist
	// Used output driver.
	output Output
	// Output volume level. 0..100
	outputVol int
	// Playback engine.
	engine *Engine
	// Channel to notify client that player state has been changed.
	events chan Event
}

func New(fmts []format.Format, output Output) *Player {
	p := &Player{
		plists:    make(map[string]*Playlist),
		curPlist:  NewPlaylist(vfsPlistName),
		output:    output,
		outputVol: 50,
		engine:    NewEngine(fmts, output),
		events:    make(chan Event, eventsChSize),
	}
	p.engine.Start()
	p.engine.SetStatusHandler(p.notifyStatus)

	return p
}

func (p *Player) Events() <-chan Event {
	return p.events
}

func (p *Player) Close() error {
	return p.engine.Close()
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

	err = p.engine.Play(p.curPlist, pos)
	if err != nil {
		return err
	}

	return nil
}

func (p *Player) Stop() error {
	return p.engine.Stop()
}

func (p *Player) Pause() error {
	return p.engine.Pause()
}

func (p *Player) Next() error {
	return p.engine.Next()
}

func (p *Player) Prev() error {
	return p.engine.Prev()
}

func (p *Player) Seek(pos int, rel bool) error {
	return p.engine.Seek(pos, rel)
}

func (p *Player) Volume() int {
	return p.outputVol
}

func (p *Player) SetVolume(vol int, rel bool) error {
	if rel {
		vol = max(0, min(100, p.outputVol+vol))
	}
	err := p.engine.Volume(vol)
	if err != nil {
		return err
	}
	p.outputVol = vol
	p.notifyStatus(p.Status())

	return nil
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
	// p.notify(PlaylistEvent, plist)

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
	p.notify(&PlistCreateEvent{name})

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
		err := p.engine.Stop()
		if err != nil {
			return err
		}
		p.curPlist = nil
	}

	p.notify(&PlistDeleteEvent{name})

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
	p.notify(&PlistRenameEvent{from, to})

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
	return p.engine.Status()
}

func (p *Player) notifyStatus(s *Status) {
	e := &StatusEvent{
		State:  s.State,
		Volume: p.Volume(),
	}
	if s.State != StateStopped {
		e.State = s.State
		e.Plist = s.Plist
		e.PlistPos = s.PlistPos
		e.Track = s.Plist.Get(s.PlistPos)
		e.TrackPos = s.Pos
	}
	p.notify(e)
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
		// TODO: p.engine.SetPlaylist(pl)
	}
}

func (p *Player) notify(e Event) {
	if len(p.events) < eventsChSize {
		p.events <- e
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
