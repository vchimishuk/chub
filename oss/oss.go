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

package oss

// #include <stdlib.h>
// #include <sys/soundcard.h>
// #include "oss.h"
import "C"

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

type Oss struct {
	fd     int
	rate   int
	chans  int
	paused bool
}

func New() *Oss {
	return &Oss{
		rate:  41000,
		chans: 2,
	}
}

func (o *Oss) Open() error {
	dev := devFile()
	if dev == "" {
		return errors.New("no OSS device found")
	}

	cdev := C.CString(dev)
	defer C.free(unsafe.Pointer(cdev))
	fd, err := C.oss_open(cdev)
	if err != nil {
		return fmt.Errorf("failed to open OSS device: %w", err)
	}
	o.fd = int(fd)

	// TODO: Try native order here, ALSA and FFMpeg.
	e, err := C.oss_format(C.int(o.fd), C.AFMT_S16_NE)
	if e == -1 {
		return err
	}

	o.SetSampleRate(o.rate)
	o.SetChannels(o.chans)
	o.paused = false

	return nil
}

func (o *Oss) SetSampleRate(rate int) error {
	e, err := C.oss_sample_rate(C.int(o.fd), C.int(rate))
	if e == -1 {
		return err
	}
	o.rate = rate

	return nil
}

func (o *Oss) SetChannels(chans int) error {
	e, err := C.oss_channels(C.int(o.fd), C.int(chans))
	if e == -1 {
		return err
	}
	o.chans = chans

	return nil
}

func (o *Oss) Write(buf []byte) (written int, err error) {
	if len(buf) == 0 {
		return 0, nil
	}

	n, err := C.oss_write(C.int(o.fd), unsafe.Pointer(&buf[0]),
		C.int(len(buf)))
	if n == -1 {
		return 0, err
	}

	return int(n), nil
}

func (o *Oss) Flush() error {
	return nil
}

func (o *Oss) Pause() error {
	var err error

	if o.paused {
		err = o.Open()
		o.paused = false
	} else {
		err = o.Close()
		o.paused = true
	}
	if err != nil {
		return err
	}

	return nil
}

func (o *Oss) Close() error {
	o.Flush()
	e, err := C.oss_close(C.int(o.fd))
	if e == -1 {
		return err
	}
	o.fd = 0

	return nil
}

func devFile() string {
	if fileExists("/dev/dsp") {
		return "/dev/dsp"
	}
	if fileExists("/dev/dsp0") {
		return "/dev/dsp0"
	}

	return ""
}

func fileExists(name string) bool {
	_, err := os.Stat("/dev/dsp")

	return err == nil
}
