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
	"time"

	"github.com/vchimishuk/chub/vfs"
)

type command int

const (
	cmdNext command = iota
	cmdPause
	cmdPlay
	cmdPlist
	cmdPrev
	cmdQuit
	cmdStop
)

type message struct {
	cmd  command
	args []interface{}
}

type state int

const (
	stateStopped state = iota
	statePlaying
	statePaused
)

type playingThread struct {
	decoders     map[string]func() Decoder
	output       Output
	plist        []*vfs.Track
	pos          int
	msgChan      chan *message
	state        state
	decoder      Decoder
	bufAvailable chan struct{}
}

func newPlayingThread(decoders map[string]func() Decoder,
	output Output) *playingThread {

	return &playingThread{
		decoders: decoders,
		output:   output,
		pos:      -1,
		msgChan:  make(chan *message),
		state:    stateStopped,
	}
}

func (pt *playingThread) Start() {
	go pt.loop()
}

func (pt *playingThread) Stop() {
	pt.msgChan <- &message{cmd: cmdStop, args: []interface{}{}}
	// Wait till loop() closes a channel before exit.
	<-pt.msgChan
}

func (pt *playingThread) Play(plist []*vfs.Track, pos int) {
	pt.msgChan <- &message{cmd: cmdPlay, args: []interface{}{plist, pos}}
}

func (pt *playingThread) Pause() {
	pt.msgChan <- &message{cmd: cmdPause, args: []interface{}{}}
}

func (pt *playingThread) Next() {
	pt.msgChan <- &message{cmd: cmdNext, args: []interface{}{}}
}

func (pt *playingThread) Prev() {
	pt.msgChan <- &message{cmd: cmdPrev, args: []interface{}{}}
}

func (pt *playingThread) SetPlaylist(plist []*vfs.Track) {
	pt.msgChan <- &message{cmd: cmdPlist, args: []interface{}{plist}}
}

func (pt *playingThread) loop() {
	var quit bool = false
	var buf []byte

	// TODO: Close decoder on prev/next, stop, etc.
	for !quit {
		// Sleep select. Wait output to be ready to consume new portion
		// of PCM data. Or handle some command if any arrives.
		select {
		case msg := <-pt.msgChan:
			switch msg.cmd {
			case cmdPlist:
				pt.setPlaylist(msg.args[0].([]*vfs.Track))
				if pt.pos == -1 {
					pt.stop()
				}
			case cmdPlay:
				pt.setPlaylist(msg.args[0].([]*vfs.Track))
				pt.play(msg.args[1].(int))
			case cmdQuit:
				quit = true
				fallthrough
			case cmdStop:
				pt.stop()
			case cmdPause:
				if pt.state == statePlaying {
					pt.stopBufAvailableChecker()
					pt.output.Pause()
					pt.state = statePaused
				} else if pt.state == statePaused {
					pt.startBufAvailableChecker()
					pt.output.Pause()
					pt.state = statePlaying
				}
			case cmdNext, cmdPrev:
				var pos int

				if msg.cmd == cmdNext {
					pos = pt.pos + 1
				} else {
					pos = pt.pos - 1
				}

				pt.play(pos)
			default:
				panic("unsupported command")
			}
		case <-pt.bufAvailable:
			// Output buffer is available now for some new portion
			// of decoded data. Just wake up and decode some.
		}

		if pt.state == statePlaying {
			// TODO: Log errors in debug mode here.
			size, _ := pt.output.AvailUpdate()
			// Do not allocate new buffer if old one is big enough.
			if cap(buf) >= size {
				buf = buf[:size]
			} else {
				buf = make([]byte, size)
			}

			// TODO: Stopped here.
			//
			// Stop this track decoding if track time exceeded: close
			// decoder and move to next track. But if next track lives
			// in the same file decoder can be reused and position
			// can be simply seeked or even noop (neighbour tracks).

			read, _ := pt.decoder.Read(buf)
			if read == 0 {
				pt.play(pt.pos + 1)
			} else {
				// TODO: If wrote not all data?
				pt.output.Write(buf)
				// TODO: Log when read != wrote in debug mode.
			}
		}
	}

	close(pt.msgChan)
}

func (pt *playingThread) play(pos int) {
	if pos < 0 {
		pos = len(pt.plist) - 1
	} else if pos >= len(pt.plist) {
		pos = 0
	}

	if pt.state != stateStopped {
		// TODO: Empty output.
		pt.stopBufAvailableChecker()
		pt.decoder.Close()
		pt.output.Close()
		pt.state = stateStopped
	}

	pth := pt.plist[pos].Path
	df := pt.decoders[pth.Ext()]
	if df == nil {
		// TODO: Skip this track and try next one.
		panic("TODO:")
	}
	decoder := df()
	err := decoder.Open(pth.Full())
	if err != nil {
		// TODO: Skip this track and try next one.
		panic("TODO:")
	}

	// TODO: Reset hw params on track change if needed.
	// TODO: Check errors.
	pt.output.Open()
	pt.output.SetSampleRate(44100)
	pt.output.SetChannels(2)

	pt.pos = pos
	pt.decoder = decoder
	pt.startBufAvailableChecker()
	pt.state = statePlaying
}

func (pt *playingThread) stop() {
	if pt.state != stateStopped {
		pt.stopBufAvailableChecker()
		pt.output.Close()
		pt.decoder.Close()
		pt.plist = nil
		pt.pos = -1
		pt.state = stateStopped
	}
}

func (pt *playingThread) setPlaylist(plist []*vfs.Track) {
	// Try to find current track in new playlist.
	if pt.state != stateStopped {
		cur := pt.plist[pt.pos]
		pt.pos = -1
		for i, t := range plist {
			// Yes, compare pointers. It helps us handle
			// track duplications in the playlist.
			if t == cur {
				pt.pos = i
				break
			}
		}
	}

	pt.plist = cloneTracks(plist)
}

func (pt *playingThread) startBufAvailableChecker() {
	pt.bufAvailable = make(chan struct{})
	go bufAvailableChecker(pt.output, pt.bufAvailable)
}

func (pt *playingThread) stopBufAvailableChecker() {
	pt.bufAvailable <- struct{}{}
	close(pt.bufAvailable)
}

// buffAvailableChecker monitors output buffer and signals via the given
// channel when there is some free space available in the buffer, so player
// can decode next piece of audio data and write it into the buffer.
func bufAvailableChecker(output Output, ch chan struct{}) {
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

func cloneTracks(tracks []*vfs.Track) []*vfs.Track {
	s := make([]*vfs.Track, len(tracks))
	copy(s, tracks)

	return s
}
