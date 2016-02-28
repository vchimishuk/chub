package main

import (
	"fmt"

	"github.com/vchimishuk/chub/mp3/libmad"
)

func main() {
	d, err := libmad.New("/home/viacheslav/documents/music/Heavy Metal/Bon Jovi/1984 - Bon Jovi/01 - Runaway.mp3") // 226 (3:46) | 230 (3:50)
	// d, err := libmad.New("/home/viacheslav/documents/music/Heavy Metal/Bon Jovi/1984 - Bon Jovi/02 - Roulette.mp3") // 275 (4:35) | (4:40)
	// d, err := libmad.New("/home/viacheslav/documents/music/Heavy Metal/Bon Jovi/1984 - Bon Jovi/03 - She Don't Know Me.mp3") // 234 (3:54) | (03:58)
	// d, err := libmad.New("/home/viacheslav/documents/music/Heavy Metal/Bon Jovi/1984 - Bon Jovi/04 - Shot Through The Heart.mp3") // 253 (4:13) | (04:18)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(d.Length())
		d.Close()
	}
}
