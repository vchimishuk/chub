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

package player

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vchimishuk/chub/vfs"
)

const (
	vfsPlaylistName = "*vfs*"
)

type Format interface {
	Extensions() []string
	Decoder() Decoder
}

type state int

const (
	stateStopped state = iota
	statePlaying
	statePaused
)

type Player struct {
	// Mutex guards plists, vfsPlist and curPlist fields.
	// Any manipulation on that fields should be guarded with this mutex.
	plistsMu sync.RWMutex
	plists   map[string]*Playlist
	vfsPlist *Playlist
	curPlist *Playlist
	msgChan  chan *message
	decoders map[string]func() Decoder
	output   Output
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
		msgChan:  make(chan *message),
		decoders: decoders,
		output:   output,
	}

	go p.run()

	return p
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
			} else if pos == -1 && &path == &t.Path {
				pos = i
			}
			p.vfsPlist.Append(t)
		}
	}
	p.curPlist = p.vfsPlist
	go p.sendMsg(cmdPlay, p.curPlist.Tracks(), pos)

	return nil
}

func (p *Player) Stop() error {
	return p.sendMsg(cmdStop)
}

func (p *Player) Pause() error {
	return p.sendMsg(cmdPause)
}

func (p *Player) Next() error {
	return p.sendMsg(cmdNext)
}

func (p *Player) Prev() error {
	return p.sendMsg(cmdPrev)
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
		p.sendMsg(cmdPlist, pl.Tracks())
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
		p.sendMsg(cmdPlist, pl.Tracks())
	}

	return nil
}

func (p *Player) Create(plist string) error {
	p.plistsMu.Lock()
	defer p.plistsMu.Unlock()

	_, err := p.playlist(plist)
	if err == nil {
		return errors.New("playlist already exists")
	}
	p.plists[plist] = NewPlaylist(plist)

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
		p.sendMsg(cmdStop)
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

// playlist returns existing playlist by name.
// This method must be guarded by player global lock.
func (p *Player) playlist(name string) (*Playlist, error) {
	if plist, ok := p.plists[name]; ok {
		return plist, nil
	} else {
		return nil, errors.New("playlist not found")
	}
}

// sendMsg send message to playing routine.
func (p *Player) sendMsg(cmd command, args ...interface{}) error {
	msg := newMessage(cmd, args)
	p.msgChan <- msg
	return msg.GetResult()
}

// run runs decode -> output loop, -- heart of playing process.
func (p *Player) run() {
	var plist []*vfs.Track
	var pos int = -1
	var buf []byte
	var bufAvailable chan bool
	var decoder Decoder
	var st state = stateStopped
	var quit bool = false

	startBufAvailableChecker := func(output Output) chan bool {
		ch := make(chan bool)
		go bufAvailableChecker(output, ch)
		return ch
	}
	getDecoder := func() Decoder {
		pth := plist[pos].Path
		d := p.decoder(pth)
		if d == nil {
			// TODO: Log errors.New("not supported format")
			panic(nil)
		}
		if err := d.Open(pth.Full()); err != nil {
			// TODO: Log and skip this track.
			panic(err.Error())
		}

		return d
	}
	changeTrack := func(next bool) {
		if next {
			pos++
			if pos >= len(plist) {
				pos = 0
			}
		} else {
			pos--
			if pos < 0 {
				pos = len(plist) - 1
			}
		}
	}

	// TODO: Close decoder on prev/next, stop, etc.
	for !quit {
		// Sleep select. Wait output to be ready to consume new portion
		// of PCM data. Or handle some command if any arrives.
		select {
		case msg := <-p.msgChan:
			switch msg.cmd {
			case cmdPlist:
				track := plist[pos]
				pos = -1
				plist = cloneTracks(msg.args[0].([]*vfs.Track))

				for i, t := range plist {
					if t == track {
						pos = i
						break
					}
				}
				if pos == -1 {
					// TODO: Stop if not.
					panic("TODO: Stop if not")
				}
			case cmdPlay:
				plist = cloneTracks(msg.args[0].([]*vfs.Track))
				pos = msg.args[1].(int)

				fmt.Println("pt.run(): cmdPlay received")
				if st != stateStopped {
					bufAvailable <- true
					decoder.Close()
					p.output.Close()
					st = stateStopped
				}

				p.output.Open()
				// TODO: Reset hw params on track change if needed.
				p.output.SetSampleRate(44100)
				p.output.SetChannels(2)

				decoder = getDecoder()
				if decoder == nil {
					// TODO: Stop. falltrhrough?
				}
				bufAvailable = startBufAvailableChecker(p.output)
				st = statePlaying
			case cmdQuit:
				quit = true
				fallthrough
			case cmdStop:
				if st != stateStopped {
					bufAvailable <- true
					decoder.Close()
					p.output.Close()
					plist = nil
					pos = -1
				}
				st = stateStopped
			case cmdPause:
				if st == statePlaying {
					bufAvailable <- true
					p.output.Pause()
					st = statePaused
				} else if st == statePaused {
					bufAvailable = startBufAvailableChecker(p.output)
					p.output.Pause()
					st = statePlaying
				}
			case cmdNext, cmdPrev:
				if st != stateStopped {
					bufAvailable <- true
					decoder.Close()
					// TODO: Reset output params?
					changeTrack(msg.cmd == cmdNext)
					decoder = getDecoder()
					bufAvailable = startBufAvailableChecker(p.output)
					st = statePlaying
				}
			default:
				panic(nil)
			}

			// TODO: Do we need some error here? Do we need this res?
			msg.SendResult(nil)
		case <-bufAvailable:
			// Output buffer is available now for some new portion
			// of decoded data. Just wake up and decode some.
		}

		if st == statePlaying {
			// TODO: Log errors in debug mode here.
			size, _ := p.output.AvailUpdate()
			// Do not allocate new buffer if old one is big enough.
			if cap(buf) >= size {
				buf = buf[:size]
			} else {
				buf = make([]byte, size)
			}
			read, _ := decoder.Read(buf)
			if read == 0 {
				decoder.Close()
				changeTrack(true)
				decoder = getDecoder()
				if decoder == nil {
					// TODO: Stop.
					panic(nil)
				}
				// TODO: Reset output params?
			} else {
				// TODO: If wrote not all data?
				p.output.Write(buf)
				// TODO: Log when read != wrote in debug mode.
			}
		}
	}

	close(p.msgChan)
}

func (p *Player) decoder(path *vfs.Path) Decoder {
	if d, ok := p.decoders[path.Ext()]; ok {
		return d()
	} else {
		return nil
	}
}

// buffAvailableChecker monitors output buffer and signals via the given
// channel when there is some free space available in the buffer, so player
// can decode next piece of audio data and write it into the buffer.
func bufAvailableChecker(output Output, ch chan bool) {
	fmt.Println("bufAvailableChecker(): started")

	for {
		ready, err := output.Wait(100)
		if err != nil {
			// Sometimes Wait failed, I don't know why.
			// so just wait some time and retry.
			// TODO: Add error handling into alsalib wrapper
			//       and maybe we will have some more
			//       sensible error here.
			time.Sleep(100 * time.Millisecond)
		} else if ready {
			select {
			case ch <- true:
			case <-ch:
				// Player stopped or paused and do not
				// interested in our notifications any more.
				fmt.Println("bufAvailableChecker(): finished")
				return
			}
		}
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

func cloneTracks(tracks []*vfs.Track) []*vfs.Track {
	s := make([]*vfs.Track, len(tracks))
	copy(s, tracks)

	return s
}
