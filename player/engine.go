// Copyright 2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/vchimishuk/chub/csync"
	"github.com/vchimishuk/chub/csync/job"
	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/logger"
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
	cmdPrev
	cmdSeek
	cmdStatus
	cmdStop
	cmdVolume
)

type message struct {
	cmd  command
	args []any
}

// Engine is a core of playing process. It manages decoding source file
// and pass decoded samples to the output driver.
type Engine struct {
	// Formats for every supported file extension.
	fmts map[string]format.Format
	// Active output.
	output Output
	// Output volume level.
	outputVol int
	// Active decoder.
	decoder format.Decoder
	// Active playlist.
	plist *Playlist
	// Current track number in the active playlist.
	plistPos int
	// Buffer to buffer decoded data ready for output.
	ring *BufferRing
	// Current state.
	state State
	// Notify main goroutine with user requests.
	msgs *csync.Notify
	// Decode job whic runs decodeLoop function.
	decodeJob job.Job
	// Output job which runs outputLoop function.
	outputJob job.Job
	// Mutex guards status (stPlistPos and stTrackPos) fields below.
	stMutex sync.Mutex
	// Currently playing track index in the playlist.
	stPlistPos int
	// Currently playing track time position.
	stTrackPos int
	// Callback to notify Player about playback changes.
	statusHandler func(*Status)
}

func NewEngine(fmts []format.Format, output Output) *Engine {
	fm := map[string]format.Format{}
	for _, f := range fmts {
		for _, e := range f.Extensions() {
			fm[strings.ToLower(e)] = f
		}
	}

	return &Engine{
		fmts:   fm,
		output: output,
		ring:   NewBufferRing(4096, 256),
		state:  StateStopped,
		msgs:   csync.NewNotify(),
	}
}

func (e *Engine) Start() {
	go e.run()
}

func (e *Engine) Close() error {
	return e.cmd(cmdClose, nil)
}

func (e *Engine) Stop() error {
	return e.cmd(cmdStop, nil)
}

func (e *Engine) Next() error {
	return e.cmd(cmdNext, nil)
}

func (e *Engine) Prev() error {
	return e.cmd(cmdPrev, nil)
}

func (e *Engine) Pause() error {
	return e.cmd(cmdPause, nil)
}

func (e *Engine) Play(plist *Playlist, pos int) error {
	return e.cmd(cmdPlay, []any{plist, pos})
}

func (e *Engine) Seek(pos int, rel bool) error {
	return e.cmd(cmdSeek, []any{pos, rel})
}

func (e *Engine) Status() *Status {
	s := <-e.msgs.Send(&message{cmd: cmdStatus})
	return s.(*Status)
}

func (e *Engine) Volume(vol int) error {
	return e.cmd(cmdVolume, []any{vol})
}

func (e *Engine) SetStatusHandler(h func(*Status)) {
	e.statusHandler = h
}

func (e *Engine) cmd(c command, args []any) error {
	r := <-e.msgs.Send(&message{cmd: c, args: args})
	if r == nil {
		return nil
	}

	return r.(error)
}

// run executes messages handling loop and manages decoding & output goroutines.
func (e *Engine) run() {
	var quit bool = false

	for !quit {
		var decodeDone <-chan error
		if e.decodeJob != nil {
			decodeDone = e.decodeJob.WaitChan()
		}
		var outputDone <-chan error
		if e.outputJob != nil {
			outputDone = e.outputJob.WaitChan()
		}

		select {
		case m := <-e.msgs.WaitChan():
			msg := m.Data.(*message)
			switch msg.cmd {
			case cmdPlay:
				if e.state != StateStopped {
					e.stop()
				}
				m.Result <- e.play(msg.args[0].(*Playlist),
					msg.args[1].(int), 0)
				e.emitStatus()
			case cmdClose:
				m.Result <- e.stop()
				quit = true
				e.emitStatus()
			case cmdStop:
				var err error
				if e.state != StateStopped {
					err = e.stop()
					e.emitStatus()
				}
				m.Result <- err
			case cmdPause:
				m.Result <- e.pause()
				e.emitStatus()
			case cmdNext:
				m.Result <- e.next(false)
				e.emitStatus()
			case cmdPrev:
				m.Result <- e.prev()
				e.emitStatus()
			case cmdSeek:
				m.Result <- e.seek(msg.args[0].(int),
					msg.args[1].(bool))
			case cmdStatus:
				m.Result <- e.status()
			case cmdVolume:
				m.Result <- e.volume(msg.args[0].(int))
			default:
				panic("unsupported command")
			}
		case err := <-decodeDone:
			e.decodeJob = nil
			if err == nil {
				e.next(true)
			} else {
				logger.Error("decoding failed: %s", err)
				e.stop()
				e.emitStatus()
			}
		case err := <-outputDone:
			e.outputJob = nil
			if err == nil {
				e.stop()
			} else {
				logger.Error("output failed: %s", err)
				e.stop()
			}
			e.emitStatus()
		}
	}
}

// emitStatus notifies upper level (Player) with new playback Status.
func (e *Engine) emitStatus() {
	if e.statusHandler != nil {
		go e.statusHandler(e.status())
	}
}

// status returns current playback status.
func (e *Engine) status() *Status {
	e.stMutex.Lock()
	defer e.stMutex.Unlock()

	return &Status{
		State:    e.state,
		Plist:    e.plist,
		PlistPos: e.stPlistPos,
		Pos:      e.stTrackPos,
	}
}

// play starts playback procedure from stopped state.
// After track specified by `pos` finished playback moves to the next track
// in the playlist automatically.
func (e *Engine) play(plist *Playlist, plistPos int, trackPos int) error {
	if e.state != StateStopped {
		err := e.stop()
		if err != nil {
			return err
		}
	}

	e.plist = plist
	e.plistPos = plistPos
	e.stPlistPos = e.plistPos
	e.stTrackPos = 0

	err := e.openDecoderOutput()
	if err != nil {
		return err
	}

	if trackPos != 0 {
		t := e.plist.Get(e.plistPos)
		err := e.decoder.Seek(t.Start + trackPos)
		if err != nil {
			e.decoder.Close()
			e.decoder = nil
			return err
		}
		e.stTrackPos = trackPos
	}

	e.ring.Open()
	e.decodeJob = job.Start(e.decodeLoop)
	e.outputJob = job.Start(e.outputLoop)
	e.state = StatePlaying

	return nil
}

// Stop playback by shutting down running decode and output goroutines.
func (e *Engine) stop() error {
	var derr error
	var oerr error

	e.ring.Close(true)
	if e.decodeJob != nil {
		e.decodeJob.Wait()
		e.decodeJob = nil
	}
	if e.outputJob != nil {
		e.outputJob.Wait()
		e.outputJob = nil
	}

	if e.decoder != nil {
		derr = e.decoder.Close()
		e.decoder = nil
	}
	if e.state != StateStopped {
		oerr = e.output.Close()
	}

	e.state = StateStopped

	if derr != nil {
		return derr
	} else if oerr != nil {
		return oerr
	}

	return nil
}

func (e *Engine) pause() error {
	switch e.state {
	case StatePlaying:
		err := e.outputJob.Shutdown()
		if err != nil {
			return err
		}
		e.outputJob = nil
		err = e.output.Pause()
		if err != nil {
			return err
		}
		e.state = StatePaused
	case StatePaused:
		err := e.output.Pause()
		if err != nil {
			return err
		}
		e.outputJob = job.Start(e.outputLoop)
		e.state = StatePlaying
	}

	return nil
}

func (e *Engine) next(auto bool) error {
	if e.state == StateStopped {
		// No next track since stopped.
		return nil
	}

	if auto {
		if e.plistPos == e.plist.Len()-1 {
			// End of the playlist. Playback will be stopped by
			// outputLoop's signal.
			e.ring.Close(false)

			return nil
		}

		cur := e.plist.Get(e.plistPos)
		next := e.plist.Get(e.plistPos + 1)
		smooth := cur.Part &&
			cur.Path.File() == next.Path.File() &&
			cur.End == next.Start

		e.plistPos += 1
		if !smooth {
			err := e.decoder.Close()
			if err != nil {
				logger.Error("decoder closing faile: %s", err)
			}
			err = e.openDecoder()
			if err != nil {
				// Call stop() to try cleanup.
				e.stop()

				return err
			}
		}
		e.decodeJob = job.Start(e.decodeLoop)
	} else {
		var plistPos int

		e.stMutex.Lock()
		plistPos = e.stPlistPos
		e.stMutex.Unlock()

		if plistPos == e.plist.Len()-1 {
			// We are on the last track alread.
			return nil
		}

		err := e.stop()
		if err != nil {
			return err
		}
		err = e.play(e.plist, plistPos+1, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) prev() error {
	if e.state == StateStopped {
		return nil
	}

	var plistPos int

	e.stMutex.Lock()
	plistPos = e.stPlistPos
	e.stMutex.Unlock()

	if plistPos == 0 {
		return nil
	}

	err := e.stop()
	if err != nil {
		return err
	}
	err = e.play(e.plist, plistPos-1, 0)
	if err != nil {
		return err
	}

	return nil
}

// Change current track playback position.
func (e *Engine) seek(pos int, rel bool) error {
	// TODO: Current seek() implementation is very very slow. We need more
	//       sophisticated implementation instead. Is should not re-open
	//       decoder and output.
	// TODO: Support seek when StatePaused.

	if e.state != StatePlaying {
		return nil
	}

	err := e.stop()
	if err != nil {
		return err
	}

	t := e.plist.Get(e.stPlistPos)
	plistPos := e.stPlistPos
	var trackPos int
	if rel {
		trackPos = e.stTrackPos + pos
	} else {
		trackPos = pos
	}
	trackPos = max(0, trackPos)
	if t.Part {
		trackPos = min(t.End, trackPos)
	}

	err = e.play(e.plist, plistPos, trackPos)
	if err != nil {
		return err
	}

	return nil
}

// Set current volume.
func (e *Engine) volume(vol int) error {
	e.outputVol = vol

	if e.state == StatePlaying {
		return e.output.SetVolume(vol)
	}

	return nil
}

// Open decoder & output drivers and prepare them for the playback process.
func (e *Engine) openDecoderOutput() error {
	err := e.openDecoder()
	if err != nil {
		return err
	}

	err = e.output.Open()
	if err != nil {
		e.decoder.Close()
		e.decoder = nil
		return err
	}

	e.output.SetSampleRate(e.decoder.SampleRate())
	e.output.SetChannels(e.decoder.Channels())
	e.output.SetVolume(e.outputVol)

	return nil
}

// Open decoder for the current playlist and track.
func (e *Engine) openDecoder() error {
	t := e.plist.Get(e.plistPos)
	ext := strings.ToLower(t.Path.Ext())
	f := e.fmts[ext]
	if f == nil {
		return fmt.Errorf("unsupported format: %s", ext)
	}

	d, err := f.Decoder(t.Path.File())
	if err != nil {
		return err
	}
	if t.Part {
		err := d.Seek(t.Start)
		if err != nil {
			d.Close()
			return err
		}
	}

	e.decoder = d

	return nil
}

// decodeLoop runs a blockng IO loop that transfers data from the initialized
// decoder to the BufferRing from which data is read by the output goroutine
// in its turn.
func (e *Engine) decodeLoop(close <-chan any) error {
	var n int
	var err error
	// Time when to stop decoding to give engine a chance to change
	// its state, to make it switch to the next track.
	// Relevant only for partial tracks.
	var end int = -1
	t := e.plist.Get(e.plistPos)
	if t.Part {
		end = t.End
	}

loop:
	for {
		select {
		case <-close:
			break loop
		default:
		}

		time := e.decoder.Time()
		if end != -1 && time >= end {
			// End of partial track.
			break
		}

		buf := e.ring.PeekFree()
		if buf == nil {
			// Ring has been closed -- stop request.
			break
		}

		buf.plistPos = e.plistPos
		buf.trackPos = time - t.Start
		n, err = e.decoder.Read(buf.data[0:cap(buf.data)])
		if err != nil {
			// Decoding error -- return the error.
			break
		}
		if n == 0 {
			// Simply exit -- end of the track.
			break
		}
		buf.data = buf.data[0:n]
		e.ring.Offer(buf)
	}

	return err
}

// outputLoop runs a blocking IO loop that transfers data from buffers cache
// to the output driver.
func (e *Engine) outputLoop(close <-chan any) error {
	var curTrack int = -1
	var err error

loop:
	for {
		select {
		case <-close:
			break loop
		default:
		}

		buf := e.ring.Peek()
		if buf == nil {
			// No more data, end of the track
			// or stop request.
			break
		}

		e.stMutex.Lock()
		e.stPlistPos = buf.plistPos
		e.stTrackPos = buf.trackPos
		e.stMutex.Unlock()

		if curTrack != -1 && curTrack != buf.plistPos {
			// Emit status on automatic track change.
			e.emitStatus()
		}
		curTrack = buf.plistPos

		err = writeAll(e.output, buf.data)
		if err != nil {
			break
		}
		e.ring.OfferFree(buf)
	}

	return err
}

// writeAll writes all bytes in the given buffer into the writer
// performing multiple Write() calls if needed.
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
