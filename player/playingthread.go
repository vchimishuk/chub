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
	"io"
	"strings"
	"time"

	"github.com/vchimishuk/chub/csync"
	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/math"
)

type State int

func (s State) String() string {
	switch s {
	case StatePaused:
		return "paused"
	case StatePlaying:
		return "playing"
	case StateStopped:
		return "stopped"
	default:
		panic("unsupported state")
	}
}

const (
	StatePaused State = iota
	StatePlaying
	StateStopped
)

type Status struct {
	State    State
	Plist    *Playlist
	PlistPos int
	Pos      int
}

type command int

const (
	cmdClose command = iota
	cmdNext
	cmdPause
	cmdPlay
	cmdPlist
	cmdPrev
	cmdSeek
	cmdStatus
	cmdStop
)

type message struct {
	cmd  command
	args []interface{}
}

type playingThread struct {
	// Formats for every supported file extension.
	fmts map[string]format.Format
	// Active output.
	output Output
	// Active decoder.
	decoder format.Decoder
	// Active playlist.
	plist *Playlist
	// Current track number in the active playlist.
	pos int
	// Notify worker for client requests.
	workerNotify *csync.Notify
	// Current state.
	state State
	// Channel to notify worker that output is ready to consume
	// new portion of decoded data.
	bufAvail      chan struct{}
	statusHandler func(*Status)
}

func newPlayingThread(fmts []format.Format, output Output) *playingThread {
	fm := map[string]format.Format{}
	for _, f := range fmts {
		for _, e := range f.Extensions() {
			fm[strings.ToLower(e)] = f
		}
	}

	return &playingThread{
		fmts:         fm,
		output:       output,
		pos:          -1,
		workerNotify: csync.NewNotify(),
		bufAvail:     make(chan struct{}),
		state:        StateStopped,
	}
}

func (pt *playingThread) SetStatusHandler(h func(*Status)) {
	pt.statusHandler = h
}

func (pt *playingThread) Start() {
	go pt.worker()
}

func (pt *playingThread) Seek(pos int, rel bool) {
	pt.workerNotify.Send(&message{cmd: cmdSeek,
		args: []interface{}{pos, rel}})
}

func (pt *playingThread) Stop() {
	pt.workerNotify.Send(&message{cmd: cmdStop})
}

func (pt *playingThread) Close() {
	// Wait confirmation from worker to be sure
	// its stop process is finished.
	<-pt.workerNotify.Send(&message{cmd: cmdClose})
}

func (pt *playingThread) Play(plist *Playlist, pos int) {
	msg := &message{cmd: cmdPlay, args: []interface{}{plist, pos}}
	pt.workerNotify.Send(msg)
}

func (pt *playingThread) Pause() {
	pt.workerNotify.Send(&message{cmd: cmdPause})
}

func (pt *playingThread) Next() {
	pt.workerNotify.Send(&message{cmd: cmdNext})
}

func (pt *playingThread) Prev() {
	pt.workerNotify.Send(&message{cmd: cmdPrev})
}

func (pt *playingThread) SetPlaylist(plist *Playlist) {
	msg := &message{cmd: cmdPlist, args: []interface{}{plist}}
	pt.workerNotify.Send(msg)
}

func (pt *playingThread) Status() *Status {
	s := <-pt.workerNotify.Send(&message{cmd: cmdStatus})
	return s.(*Status)
}

func (pt *playingThread) worker() {
	// Limit max buffer size because output.Write takes too long on large
	// buffer, which cause client commands processing lag.
	const maxBufSize = 4096
	var quit bool = false
	var buf [maxBufSize]byte

	// TODO: Close decoder on prev/next, stop, etc.
	for !quit {
		// Sleep select. Wait output to be ready to consume new portion
		// of PCM data. Or handle some command if any arrives.
		select {
		case m := <-pt.workerNotify.WaitChan():
			msg := m.Data.(*message)
			switch msg.cmd {
			case cmdPlist:
				pt.setPlaylist(msg.args[0].(*Playlist))
				if pt.pos == -1 {
					pt.stop()
				}
			case cmdPlay:
				pt.setPlaylist(msg.args[0].(*Playlist))
				pt.play(msg.args[1].(int), false)
			case cmdClose:
				pt.stop()
				quit = true
				m.Result <- struct{}{}
			case cmdStop:
				pt.stop()
			case cmdPause:
				if pt.state == StatePlaying {
					pt.output.Pause()
					pt.stopBufAvailableChecker()
					pt.state = StatePaused
					pt.emitStatus()
				} else if pt.state == StatePaused {
					pt.output.Pause()
					pt.startBufAvailableChecker()
					pt.state = StatePlaying
					pt.emitStatus()
				}
			case cmdNext, cmdPrev:
				var pos int

				if msg.cmd == cmdNext {
					pos = pt.pos + 1
				} else {
					pos = pt.pos - 1
				}

				pt.play(pos, false)
			case cmdSeek:
				pt.seek(msg.args[0].(int), msg.args[1].(bool))
			case cmdStatus:
				m.Result <- pt.status()
			default:
				panic("unsupported command")
			}
		case <-pt.bufAvail:
			// Output buffer is available now for some new portion
			// of decoded data. Just wake up and decode some.
		}

		if pt.state == StatePlaying {
			size, err := pt.output.AvailUpdate()
			if err != nil {
				// TODO: Error handling.
				panic(err)
			}
			if size > 0 {
				if size > len(buf) {
					size = len(buf)
				}

				cur := pt.plist.Get(pt.pos)
				read := 0

				if !cur.Part || pt.decoder.Time() < cur.End {
					var err error
					read, err = pt.decoder.Read(buf[:size])
					if err != nil {
						// Ignore errors and treat them as
						// end of the file.
						read = 0
					}
				}
				if read == 0 {
					// TODO: Repeat support.
					if pt.pos+1 < pt.plist.Len() {
						pt.play(pt.pos+1, true)
					} else {
						pt.stop()
					}
				} else {
					err := writeAll(pt.output, buf[:read])
					if err != nil {
						// TODO: Error handling.
						panic(err)
					}
				}
			}
		}
	}
}

func (pt *playingThread) play(pos int, smooth bool) {
	if pt.plist.Len() == 0 {
		return
	}
	if pos < 0 {
		pos = pt.plist.Len() - 1
	} else if pos >= pt.plist.Len() {
		pos = 0
	}
	if pt.state == StatePlaying {
		// TODO: Flush output buffer.
		pt.stopBufAvailableChecker()
	}

	track := pt.plist.Get(pos)
	sameFile := false

	if pt.state == StatePlaying {
		cur := pt.plist.Get(pt.pos)
		sameFile = cur.Path.File() == track.Path.File()
	}

	// Do not reopen decoder if next track from the same physical file
	// as a current one.
	if !sameFile {
		if pt.state != StateStopped {
			pt.decoder.Close()
			pt.state = StateStopped
		}

		f := pt.fmts[track.Path.Ext()]
		if f == nil {
			// TODO: Skip this track and try next one.
			panic("TODO:")
		}
		d, err := f.Decoder(track.Path.File())
		if err != nil {
			pt.state = StateStopped
			// TODO: Skip this track and try next one.
			panic("TODO:")
		}
		pt.decoder = d
	}
	if track.Part {
		pt.decoder.Seek(track.Start)
	}

	if !smooth && pt.output.IsOpen() {
		pt.output.Close()
	}
	if !pt.output.IsOpen() {
		err := pt.output.Open()
		if err != nil {
			// TODO: Some adequate error handling.
			panic(err)
		}
	}

	osr := pt.output.SampleRate()
	och := pt.output.Channels()
	dsr := pt.decoder.SampleRate()
	dch := pt.decoder.Channels()
	if osr != dsr || och != dch {
		pt.output.SetSampleRate(dsr)
		pt.output.SetChannels(dch)
	}

	pt.pos = pos
	pt.state = StatePlaying
	pt.startBufAvailableChecker()
	pt.emitStatus()
}

func (pt *playingThread) seek(pos int, rel bool) {
	if pt.state == StatePlaying {
		tr := pt.plist.Get(pt.pos)
		var t int
		if rel {
			t = math.Max(tr.Start, pt.decoder.Time()+pos)
		} else {
			t = math.Max(tr.Start, tr.Start+pos)
		}

		pt.decoder.Seek(t)
		// TODO: Flush output buffer.
		pt.emitStatus()
	}
}

func (pt *playingThread) stop() {
	if pt.state != StateStopped {
		if pt.state == StatePlaying {
			pt.stopBufAvailableChecker()
		}
		pt.output.Close()
		pt.decoder.Close()
		pt.plist = nil
		pt.pos = -1
		pt.state = StateStopped
		pt.emitStatus()
	}
}

func (pt *playingThread) setPlaylist(plist *Playlist) {
	// Try to find current track in new playlist.
	if pt.state != StateStopped {
		cur := pt.plist.Get(pt.pos)
		pt.pos = -1
		for i := 0; i < plist.Len(); i++ {
			if plist.Get(i).Path.String() == cur.Path.String() {
				pt.pos = i
				break
			}
		}
	}
	if pt.pos == -1 {
		pt.stop()
	}
	pt.plist = plist
}

func (pt *playingThread) emitStatus() {
	if pt.statusHandler != nil {
		go pt.statusHandler(pt.status())
	}
}

func (pt *playingThread) status() *Status {
	s := &Status{}
	s.State = pt.state
	s.Plist = pt.plist
	s.PlistPos = pt.pos
	if s.State != StateStopped {
		t := pt.plist.Get(pt.pos)
		s.Pos = pt.decoder.Time() - t.Start
	}

	return s
}

func (pt *playingThread) startBufAvailableChecker() {
	go bufAvailChecker(pt, pt.output, pt.bufAvail)
}

func (pt *playingThread) stopBufAvailableChecker() {
	pt.bufAvail <- struct{}{}
}

// buffAvailableChecker monitors output buffer and signals via the given
// channel when there is some free space available in the buffer, so player
// can decode next piece of audio data and write it into the buffer.
func bufAvailChecker(pt *playingThread, output Output, ch chan struct{}) {
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
			case ch <- struct{}{}:
			case <-ch:
				// Player stopped or paused and do not
				// interested in our notifications any more.
				return
			}
		}
	}
}

func writeAll(w io.Writer, buf []byte) error {
	for len(buf) > 0 {
		n, err := w.Write(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
	}

	return nil
}
