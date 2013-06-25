package playingthread

import (
	"../../audio"
	"../../vfs"
	"../playlist"
	"fmt"
	"time"
)

// Size of the buffer for decoding data into.
const decoderBufferSize = 64 * 1024

// threadState type describes state of the playing thread.
type state int

const (
	// Thread is currently stopped.
	stateStopped state = iota
	// Thread is playing some track now.
	statePlaying
	// Thread is paused.
	statePaused
)

// Actual playing object.
type PlayingThread struct {
	// Current thread state.
	state state
	// Output driver.
	output audio.Output
	// Decoder driver implementation for the current playing track.
	// This is not nil only if thread in threadStatePlaying state.
	decoder audio.Decoder
	// Channel to communicate with the core routine.
	commandChan chan *command
	// Current playing playlist.
	plist *playlist.Playlist
	// Current playing track index.
	currTrack int
}

// Returns new playingThread object redy to start playback.
func New(plist *playlist.Playlist) *PlayingThread {
	thread := &PlayingThread{
		state:       stateStopped,
		output:      nil,
		decoder:     nil,
		commandChan: make(chan *command),
		plist:       plist,
		currTrack:   -1}

	return thread
}

// Run runs playing thread internal goroutine, after that PlayingThread redy
// for command processing.
func (thread *PlayingThread) Run() error {
	output, err := audio.GetOutput()
	if err != nil {
		return err
	}
	thread.output = output

	go thread.routine()

	return nil
}

// Play starts playing current playlist starting from the given track number.
func (thread *PlayingThread) Play(track int) error {
	thread.currTrack = track

	cmd := newCommand(actionPlay, nil)
	err := thread.commandDispatcher(cmd)

	return err
}

// Stop stop playing.
func (thread *PlayingThread) Stop() {
	cmd := newCommand(actionStop, nil)
	thread.commandDispatcher(cmd)
}

// Pause pauses or resumes current playback process.
func (thread *PlayingThread) Pause() {
	cmd := newCommand(actionPause, nil)
	thread.commandDispatcher(cmd)
}

// Returns true if playback is paused now.
func (thread *PlayingThread) Paused() bool {
	return thread.state == statePaused
}

// commandDispatcher is an synchronous interface for playing thread goroutine.
func (thread *PlayingThread) commandDispatcher(cmd *command) error {
	thread.commandChan <- cmd
	err := <-cmd.resultChan

	return err
}

// Returns next track for playing.
// true stop value means that there are no tracks in the playlist
// and player should be stopped.
func (thread *PlayingThread) getNextTrack() (track *vfs.Track, stop bool, err error) {
	thread.plist.Lock()
	defer thread.plist.Unlock()

	currTrack := thread.currTrack

	// Stopped state means that we are going to start playing and thread.currTrack
	// points to the track for starting playback from.
	// Non stopped state means that we have played some track and want to
	// move to next one, so currTrack should be moved forward.
	if thread.state != stateStopped {
		currTrack++
	} else {
		err := thread.output.Open()
		if err != nil {
			return nil, true, fmt.Errorf("Failed to open output driver. %s", err)
		}

		// TODO: Reset hw params on track change if needed.
		thread.output.SetSampleRate(44100)
		thread.output.SetChannels(2)
	}

	// If next track in the playlist can't be played try
	// next after it, and so forth till the end of the playlist.
	for ; currTrack < thread.plist.Len(); currTrack++ {
		track := thread.plist.Get(currTrack)

		decoder, err := audio.GetDecoder(track)
		if err != nil {
			// Try next track.
			continue
		}
		err = decoder.Open(track)
		if err != nil {
			// Try next track.
			continue
		}

		thread.currTrack = currTrack
		thread.decoder = decoder

		return track, false, nil
	}

	// No track found.
	return nil, true, nil
}

// buffAvailableChecker monitors output buffer and signals via given
// channel when there is some free space in the buffer, so we can
// decode piece of audio data and write to the buffer.
func (thread *PlayingThread) bufferAvailableChecker(ch chan bool) {
	// TODO: Add mutex to be sure that only one instance of this
	//       function exists.
	for thread.state != stateStopped {
		ready, err := thread.output.Wait(100)

		if err != nil {
			// Sometimes Wait failed, I don't know why.
			// so just wait some time and retry.
			// TODO: Add arror handling into alsalib wrapper
			//       and maybe we will have some
			//       sensible error here.
			time.Sleep(100 * time.Millisecond)
		} else if ready && thread.state == statePlaying {
			ch <- true
		} else { // Paused.
			time.Sleep(300 * time.Millisecond)
		}
	}
}

// Switch to stopped state.
func (thread *PlayingThread) stop() {
	thread.state = stateStopped
	// XXX: Wait for buffer checker mutex here if alsa driver
	//      will crash sometimes (e.g. in Wait()).
	thread.decoder.Close()
	thread.decoder = nil
	thread.output.Close()
}

// routine is a heart of the playing process.
func (thread *PlayingThread) routine() {
	// TODO: Log track management here.

	var quit bool = false
	var bufAvailable chan bool = make(chan bool)
	var buffer [decoderBufferSize]byte

	// I'm in stop state here.

	for {
		select {
		case cmd := <-thread.commandChan:
			var err error

			switch cmd.action {
			case actionPlay:
				// Let's stop first if play action arrives when we already playing.
				if thread.state != stateStopped {
					thread.stop()
				}

				track, stop, err := thread.getNextTrack()
				if err != nil || stop || track == nil {
					// Keep stay stopped.
				} else {
					thread.state = statePlaying
					go thread.bufferAvailableChecker(bufAvailable)
				}
			case actionStop:
				if thread.state != stateStopped {
					thread.stop()
				}
			case actionPause:
				thread.output.Pause()

				if thread.state == statePlaying {
					thread.state = statePaused
				} else if thread.state == statePaused {
					thread.state = statePlaying
				}
			case actionQuit:
				thread.stop()
				quit = true
			}

			cmd.resultChan <- err
		case <-bufAvailable:
			// Do nothing, just wake up.
		}

		switch thread.state {
		case statePlaying:
			// TODO: Log errors in debug mode here.
			// size, _ := thread.output.AvailUpdate()
			read, _ := thread.decoder.Read(buffer[:])
			if read == 0 {
				// End of track reached, let's move to next one.
				track, stop, err := thread.getNextTrack()
				if err != nil || stop || track == nil {
					thread.Stop()
				}
			} else {
				thread.output.Write(buffer[:read])
				// TODO: Log when read != wrote in debug mode.
			}
		case stateStopped:
			// Do nothing.
		case statePaused:
			// Do nothing here, just sleep with select untill
			// resume wakes us up.
		}

		if quit {
			return
		}
	}

	close(bufAvailable)
}
