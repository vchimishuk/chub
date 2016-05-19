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

	"github.com/vchimishuk/chub/vfs"
)

const (
	vfsPlaylistName = "*vfs*"
)

type Format interface {
	Extensions() []string
	Decoder() Decoder
}

type Player struct {
	// Mutex guards plists, vfsPlist and curPlist fields.
	// Any manipulation on that fields should be guarded with this mutex.
	plistsMu sync.RWMutex
	plists   map[string]*Playlist
	vfsPlist *Playlist
	curPlist *Playlist
	decoders map[string]func() Decoder
	output   Output
	pt       *playingThread
	notifsMu sync.RWMutex // Guards next field.
	notifs   []chan *NotifMsg
}

func New(fmts []Format, output Output) *Player {
	decoders := make(map[string]func() Decoder)
	for _, f := range fmts {
		for _, ext := range f.Extensions() {
			decoders[ext] = f.Decoder
		}
	}

	p := &Player{
		plists:   make(map[string]*Playlist),
		vfsPlist: NewPlaylist(vfsPlaylistName),
		decoders: decoders,
		output:   output,
		pt:       newPlayingThread(decoders, output),
	}

	p.pt.Start()

	return p
}

func (p *Player) AddNotifier(notif chan *NotifMsg) {
	p.notifsMu.Lock()
	defer p.notifsMu.Unlock()

	p.notifs = append(p.notifs, notif)
}

func (p *Player) RemoveNotifier(notif chan *NotifMsg) {
	p.notifsMu.Lock()
	defer p.notifsMu.Unlock()

	for i, n := range p.notifs {
		if n == notif {
			p.notifs = append(p.notifs[:i], p.notifs[i+1:]...)
			break
		}
	}
}

func (p *Player) Close() {
	p.pt.Close()
}

func (p *Player) Play(path *vfs.Path) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	// Fill VFS playlist.
	var dir *vfs.Path
	var pos int = -1
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
	p.vfsPlist.Clear()
	for i, entry := range entries {
		if !entry.IsDir() {
			t := entry.(*vfs.Track)
			// Find position to start.
			if path.IsDir() && pos == -1 {
				pos = 0
			} else if pos == -1 && *path == *t.Path {
				pos = i
			}
			p.vfsPlist.Append(t)
		}
	}
	p.curPlist = p.vfsPlist
	p.pt.Play(p.curPlist.Tracks(), pos)

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

func (p *Player) Append(plist string, path *vfs.Path) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.playlist(plist)
	if err != nil {
		return err
	}

	tracks, err := listDirRec(path)
	if err != nil {
		return err
	}
	pl.Append(tracks...)
	if pl == p.curPlist {
		p.pt.SetPlaylist(pl.Tracks())
	}

	return nil
}

func (p *Player) Clear(plist string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.playlist(plist)
	if err != nil {
		return err
	}

	pl.Clear()
	if pl == p.curPlist {
		p.pt.SetPlaylist(pl.Tracks())
	}

	return nil
}

func (p *Player) Create(plist string) error {
	p.plistsMu.Lock()
	_, err := p.playlist(plist)
	if err == nil {
		p.plistsMu.Unlock()
		return errors.New("playlist already exists")
	}
	p.plists[plist] = NewPlaylist(plist)
	p.plistsMu.Unlock()

	p.notify(PlaylistsEvent, p.Playlists())

	return nil
}

func (p *Player) Delete(plist string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.playlist(plist)
	if err != nil {
		return err
	}

	delete(p.plists, plist)
	if pl == p.curPlist {
		p.pt.Stop()
	}

	return nil
}

func (p *Player) Playlist(name string) (*Playlist, error) {
	p.plistsMu.RLock()
	defer p.plistsMu.RUnlock()

	pl, err := p.playlist(name)
	if err != nil {
		return nil, err
	}

	return pl.clone(), nil
}

func (p *Player) Rename(from string, to string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	pl, err := p.playlist(from)
	if err != nil {
		return err
	}

	pl.SetName(to)

	return nil
}

func (p *Player) Playlists() []*PlaylistInfo {
	p.plistsMu.RLock()
	defer p.plistsMu.RUnlock()

	plists := make([]*PlaylistInfo, len(p.plists))
	i := 0
	for _, pl := range p.plists {
		plists[i] = pl.info()
		i++
	}

	return plists
}

func (p *Player) notify(e Event, val interface{}) {
	p.notifsMu.RLock()
	defer p.notifsMu.RUnlock()

	for _, n := range p.notifs {
		go func() {
			n <- &NotifMsg{Event: e, Value: val}
		}()
	}
}

// playlist returns existing playlist by name.
// This method must be guarded by player global lock.
func (p *Player) playlist(name string) (*Playlist, error) {
	if plist, ok := p.plists[name]; ok {
		return plist, nil
	} else {
		return nil, errors.New("playlist not found")
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
