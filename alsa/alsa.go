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

// ALSA output driver.
package alsa

import "github.com/vchimishuk/chub/alsa/asoundlib"

// DriverName is the string name of the alsa driver.
var DriverName string = "alsa"

// Alsa aoutput driter handler structure.
type Alsa struct {
	handle *asoundlib.Handle
	open   bool
}

// New returns newly initialized alsa output driver.
func New() *Alsa {
	return &Alsa{}
}

func (a *Alsa) Name() string {
	return "ALSA"
}

func (a *Alsa) Open() error {
	a.handle = asoundlib.New()
	err := a.handle.Open("default", asoundlib.StreamTypePlayback, asoundlib.ModeBlock)
	if err != nil {
		return err
	}

	a.handle.SampleFormat = asoundlib.SampleFormatS16LE
	a.handle.SampleRate = 44100
	a.handle.Channels = 2
	a.handle.ApplyHwParams()

	a.open = true

	return nil
}

func (a *Alsa) IsOpen() bool {
	return a.open
}

func (a *Alsa) SampleRate() int {
	return a.handle.SampleRate
}

func (a *Alsa) SetSampleRate(rate int) {
	a.handle.SampleRate = rate
	a.handle.ApplyHwParams()
}

func (a *Alsa) Channels() int {
	return a.handle.Channels
}

func (a *Alsa) SetChannels(channels int) {
	a.handle.Channels = channels
	a.handle.ApplyHwParams()
}

func (a *Alsa) Wait(maxDelay int) (ok bool, err error) {
	return a.handle.Wait(maxDelay)
}

func (a *Alsa) AvailUpdate() (size int, err error) {
	return a.handle.AvailUpdate()
}

func (a *Alsa) Write(buf []byte) (written int, err error) {
	return a.handle.Write(buf)
}

func (a *Alsa) Reset() {
	a.handle.Reset()
}

func (a *Alsa) Pause() {
	a.handle.Pause()
}

func (a *Alsa) Paused() bool {
	return a.handle.Paused()
}

func (a *Alsa) Close() {
	a.handle.Close()
	a.open = false
}
