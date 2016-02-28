package main

import (
	"./alsa/asoundlib"
	"./mp3/libmad"
	"fmt"
	"time"
)

func main() {
	mp3File := "/home/viacheslav/projects/chubd_music/ACDC/01 - Baby, Please Don't Go.mp3"
	var err error

	mad, err := libmad.New(mp3File)
	if err != nil {
		panic("libmad.New() failed\n")
	}

	handle := asoundlib.New()
	err = handle.Open("default", asoundlib.StreamTypePlayback, asoundlib.ModeBlock)
	if err != nil {
		panic(fmt.Sprintf("Open failed. %s", err))
	}

	handle.SampleFormat = asoundlib.SampleFormatS16LE
	handle.SampleRate = 44100
	handle.Channels = 2
	err = handle.ApplyHwParams()
	if err != nil {
		panic(fmt.Sprintf("SetHwParams failed. %s", err))
	}

	//	for i := 0; i < 20; i++ {
	buf := make([]byte, 60208*10)

	read := mad.Read(buf)
	fmt.Printf("read: %d\n", read)

	wrote, err := handle.Write(buf[:read])

	if err != nil {
		panic(fmt.Sprintf("Writei failed. %s", err))
	} else {
		fmt.Printf(fmt.Sprintf("wrote: %d\n", wrote))
	}
	//	}

	time.Sleep(60 * time.Second)

	handle.Close()
}
