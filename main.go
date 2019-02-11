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

package main

import (
	"fmt"

	"github.com/vchimishuk/chub/alsa"
	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/format/ffmpeg"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/server/cmd"
	"github.com/vchimishuk/chub/server/notif"
	"github.com/vchimishuk/chub/vfs"
)

func main() {
	ffmpegFmt := ffmpeg.NewFormat()

	// Initialize VFS.
	// vfs.RegisterFormat(flac.Format)
	// vfs.RegisterFormat(mp3.Format)
	format.Register(ffmpegFmt)

	err := vfs.SetRoot("/home/vchimishuk/server/Heavy Metal")
	if err != nil {
		panic(err)
	}

	// fmts := []player.Format{
	// 	flac.Format,
	// 	mp3.Format,
	// }
	output := alsa.New()
	pl := player.New([]format.Format{ffmpegFmt}, output)

	notifSrv := notif.NewServer(pl)
	notifSrv.Listen("127.0.0.1", 5225)
	fmt.Println("Notification server started")
	go notifSrv.Serve()

	cmdSrv := cmd.NewServer(pl)
	cmdSrv.Listen("127.0.0.1", 5115)
	fmt.Println("Command server started")
	cmdSrv.Serve()
	fmt.Println("Command server stopped")

	notifSrv.Close()
	fmt.Println("Notification server stopped")
}
