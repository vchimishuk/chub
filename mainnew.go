package main

import (
	"fmt"
	"strconv"

	"github.com/vchimishuk/chub/mp3/id3tag"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/server"
	"github.com/vchimishuk/chub/vfs"
)

func main() {
	// Initialize VFS.
	vfs.RegisterFormat("ogg")
	vfs.RegisterFormat("mp3")
	vfs.RegisterTagReader("mp3", parseId3Tag)

	s := server.NewCmdServer(player.New())
	s.Listen("localhost", 8888)

	fmt.Println("Server started")
	s.Serve()
	fmt.Println("Server stopped")

	// TODO: Make the music play!

	// TODO: Stop player after Serve() finished.

	// p := player.New()

	// p.PlaylistCreate("test")
	// path, err := vfs.NewPath("/home/viacheslav/documents/music/")
	// if err != nil {
	// 	panic(nil)
	// }
	// p.PlaylistAppend("test", path)
	// plist := p.Test()[0]

	// fmt.Printf("len: %d\n", len(plist.Tracks()))
}

func parseId3Tag(file string) (*vfs.Tag, error) {
	id3Tag, err := id3tag.Parse(file)
	if err != nil {
		return nil, err
	}

	number, err := strconv.Atoi(id3Tag.Number())
	if err != nil {
		number = 0
	}

	tag := &vfs.Tag{
		Artist: id3Tag.Artist(),
		Album:  id3Tag.Album(),
		Title:  id3Tag.Title(),
		Number: number,
		Length: 0} // TODO:

	return tag, nil
}
