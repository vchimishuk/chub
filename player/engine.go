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

func (e *Engine) Close() {
	// Wait confirmation from the worker to be sure it done its job.
	<-e.msgs.Send(&message{cmd: cmdClose})
}

func (e *Engine) Stop() {
	e.msgs.Send(&message{cmd: cmdStop})
}

func (e *Engine) Next() {
	e.msgs.Send(&message{cmd: cmdNext})
}

func (e *Engine) Prev() {
	e.msgs.Send(&message{cmd: cmdPrev})
}

func (e *Engine) Pause() {
	e.msgs.Send(&message{cmd: cmdPause})
}

func (e *Engine) Play(plist *Playlist, pos int) {
	e.msgs.Send(&message{cmd: cmdPlay, args: []interface{}{plist, pos}})
}

func (e *Engine) Seek(pos int, rel bool) {
	e.msgs.Send(&message{cmd: cmdSeek, args: []interface{}{pos, rel}})
}

func (e *Engine) Status() *Status {
	s := <-e.msgs.Send(&message{cmd: cmdStatus})
	return s.(*Status)
}

func (e *Engine) SetStatusHandler(h func(*Status)) {
	e.statusHandler = h
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
				e.play(msg.args[0].(*Playlist),
					msg.args[1].(int))
				e.emitStatus()
			case cmdClose:
				e.stop()
				quit = true
				m.Result <- struct{}{}
				e.emitStatus()
			case cmdStop:
				if e.state != StateStopped {
					e.stop()
					e.emitStatus()
				}
			case cmdPause:
				e.pause()
				e.emitStatus()
			case cmdNext:
				// TODO: Return error to the client?
				//       Same for other commands.
				e.next(false)
				e.emitStatus()
			case cmdPrev:
				e.prev()
				e.emitStatus()
			case cmdSeek:
				e.seek(msg.args[0].(int), msg.args[1].(bool))
			case cmdStatus:
				m.Result <- e.status()
			default:
				panic("unsupported command")
			}
		case err := <-decodeDone:
			if err == nil {
				e.next(true)
			} else {
				logger.Error("decoding failed: %s", err)
				e.stop()
				e.emitStatus()
			}
		case err := <-outputDone:
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
func (e *Engine) play(plist *Playlist, pos int) error {
	if e.state != StateStopped {
		err := e.stop()
		if err != nil {
			return err
		}
	}

	e.plist = plist
	e.plistPos = pos
	e.stPlistPos = e.plistPos
	e.stTrackPos = 0

	err := e.openDecoderOutput()
	if err != nil {
		return err
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

	e.ring.Close()
	if e.decodeJob != nil {
		e.decodeJob.Wait()
		e.decodeJob = nil
	}
	if e.outputJob != nil {
		e.outputJob.Wait()
		e.outputJob = nil
	}

	if e.decoder != nil {
		// TODO: derr = e.decoder.Close()
		e.decoder.Close()
		e.decoder = nil
	}
	if e.output.IsOpen() {
		// TODO: oerr = e.output.Close()
		e.output.Close()
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
		// TODO: Handle error.
		e.outputJob.Shutdown()
		e.outputJob = nil
		// TODO: Handle error.
		e.output.Pause()
		e.state = StatePaused
	case StatePaused:
		// TODO: Handle error.
		e.output.Pause()
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

			// TODO: Marck ring somehow that this is
			//       the end of the track.
			return nil
		}

		cur := e.plist.Get(e.plistPos)
		next := e.plist.Get(e.plistPos + 1)
		smooth := cur.Path.File() == next.Path.File() &&
			cur.End == next.Start

		e.plistPos += 1
		if !smooth {
			// TODO: Handle error.
			e.decoder.Close()
			// TODO: Handle error.
			e.openDecoder()
		}
		e.decodeJob = job.Start(e.decodeLoop)
	} else {
		var plistPos int

		e.stMutex.Lock()
		plistPos = e.stPlistPos
		e.stMutex.Unlock()

		if plistPos == e.plist.Len()-1 {
			return nil
		}

		err := e.stop()
		if err != nil {
			return err
		}
		err = e.play(e.plist, plistPos+1)
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
	err = e.play(e.plist, plistPos-1)
	if err != nil {
		return err
	}

	return nil
}

// Change current track playback position.
func (e *Engine) seek(pos int, rel bool) error {
	// TODO: Support seek when StatePaused.
	// TODO: Seek must use e.stTrackPos as a current time.
	panic("TODO:")
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

loop:
	for {
		select {
		case <-close:
			break loop
		default:
		}

		buf := e.ring.PeekFree()
		if buf == nil {
			// Ring has been closed -- stop request.
			break
		}

		buf.plistPos = e.plistPos
		buf.trackPos = e.decoder.Time()
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
