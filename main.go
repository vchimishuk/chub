package main

// Test VFS.
// func main() {
// 	fs := vfs.New()
// 	entries, err := fs.List()
// 	if err != nil {
// 		fmt.Printf("%s\n", err)
// 		return
// 	}

// 	fmt.Printf("Listing of %s:\n", fs.Wd())
// 	for _, entry := range entries {
// 		if dir, ok := entry.(*vfs.Directory); ok {
// 			fmt.Printf("%s/\n", dir.Name)
// 		} else if track, ok := entry.(*vfs.Track); ok {
// 			fmt.Printf("%s - %s\n", track.Tag.Artist, track.Tag.Title)
// 		} else {
// 			panic(fmt.Sprintf("Unsupported entry %v", entry))
// 		}
// 	}
// }

// Test playlists.
import (
	"./alsa"
	"./audio"
	"./mp3"
	"./ogg"
	"./player"
	"time"
	//	"./player/playlist"
	"./logger"
	"./vfs"

	"./server"
	"fmt"
)

// func getTrack() *vfs.Track {
// 	fs := vfs.New()
// 	entries, err := fs.Ls()
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, e := range entries {
// 		if track, ok := e.(*vfs.Track); ok {
// 			return track
// 		}
// 	}

// 	return nil
// }

func main() {
	ntfServer := server.NewNotificationServer()
	err := ntfServer.Serve("localhost", 1489)
	if err != nil {
		fmt.Printf("Notification server Serve failed. %s\n", err)
		return
	}

	pl := player.New()
	pl.SetStateListener(ntfServer)

	err = pl.Run()
	if err != nil {
		fmt.Printf("Failed to run player. %s\n", err)
		return
	} else {
		fmt.Printf("main: Player run.\n")
	}

	e := server.NewCommandServer(pl).Serve("localhost", 1488)
	if e != nil {
		logger.Error("Serve failed. %s", e)
	}

	time.Sleep(10000 * time.Second)

	// [BEGIN] Logger testing.
	// logger.Debug("%s", "Some debug level message.")
	// logger.Info("%s", "Some info level message.")
	// logger.Warning("%s", "Some warning level message.")
	// logger.Error("%s", "Some error level message.")
	// logger.Fatal("%s", "Some fatal level message.")
	// [END] Logger testing.

	// err := pl.AddPlaylist("foo")
	// if err != nil {
	// 	fmt.Printf("ERROR: %s\n", err)
	// }
	// pl.AppendTrack("foo", getTrack())
	// pl.AppendTrack("foo", getTrack())
	// pl.AppendTrack("foo", getTrack())

	// lists := pl.Playlists()

	// fmt.Printf("playlists: %V\n", lists)

	// err = pl.PlayTrack(getTrack())
	// if err != nil {
	// 	fmt.Printf("ERROR: %s\n", err)
	// } else {
	// 	fmt.Printf("main: Playback started.\n")
	// }

	// lists = pl.Playlists()

	// fmt.Printf("playlists: %V\n", lists)

	// { 1: Test PlayTrack, Pause and Stop.

	// fmt.Println("Start playing.")
	// pl.PlayTrack(getTrack())

	// [BEGIN] Testing track changing.
	fmt.Printf("main: Sleeping.\n")
	time.Sleep(10 * time.Second)
	fmt.Printf("Main woke up.\n")

	// [END] Testing track changing.

	// [BEGIN] Testing playback control.

	// time.Sleep(1 * time.Second)
	// pl.Pause()
	// fmt.Printf("main: Paused.\n")
	// fmt.Printf("Paused(): %v\n", pl.Paused())

	// time.Sleep(1 * time.Second)
	// pl.Pause()
	// fmt.Printf("main: Resumed.\n")
	// fmt.Printf("Paused(): %v\n", pl.Paused())

	// time.Sleep(1 * time.Second)
	// pl.Stop()
	// fmt.Printf("main: Stopped.\n")

	// time.Sleep(1 * time.Second)
	// pl.PlayTrack(getTrack())
	// fmt.Printf("main: Playing again.\n")

	// time.Sleep(1 * time.Second)
	// pl.Pause()
	// fmt.Printf("main: Paused.\n")
	// fmt.Printf("Paused(): %v\n", pl.Paused())

	// time.Sleep(1 * time.Second)
	// pl.Pause()
	// fmt.Printf("main: Resumed.\n")
	// fmt.Printf("Paused(): %v\n", pl.Paused())

	// time.Sleep(1 * time.Second)
	// pl.Stop()
	// fmt.Printf("main: Stopped.\n")

	// fmt.Printf("Sleeping before exit.\n")
	// time.Sleep(3 * time.Second)

	// [END] Testing playback control.
}

// Test audio package.
// import (
// 	"fmt"
// 	"./audio"
// 	"./mp3"
// )
// func main() {
// 	fmt.Printf("%v\n", audio.Foo())
// }

// func main() {
// 	// testVfs()
// 	// testPlaylists()
// 	testAudio()

// 	// // TODO: Parse command line parameters.

// 	// // TODO: Daemonize itself.
// 	// //       daemonize()

// 	// host := "127.0.0.1"
// 	// port := 8888

// 	// srv, err := server.NewTCPServer(host, port)
// 	// if err != nil {
// 	// 	fmt.Printf("Failed to create server. %s", err)
// 	// 	os.Exit(1)
// 	// }

// 	// // Run listening loop.
// 	// srv.SetConnectionHandler(new(protocol.ConnectionHandler))
// 	// go srv.Serve()

// 	// // On SIGTERM received we have to close all client connections
// 	// // and then exit. So we loop till this signal will be recieved.
// 	// for {
// 	// 	sig := (<-signal.Incoming).(os.UnixSignal) // XXX: Will works in windows? No way!
// 	// 	sigNum := int32(sig)

// 	// 	if sigNum == SigTerm {
// 	// 		break
// 	// 	}
// 	// }

// 	// p := player.New()
// 	// p.Run()

// 	// p.Command(player.CMD_PLAYLISTS_ADD, "foo")
// 	// p.Command(player.CMD_PLAYLISTS_ADD, "bar")
// 	// p.Command(player.CMD_PLAYLISTS_ADD, "*baz*")
// 	// resp, _ := p.Command(player.CMD_PLAYLISTS_LIST)
// 	// playlists := resp.([]*player.Playlist)

// 	// for _, playlist := range playlists {
// 	// 	fmt.Printf("Playlist: %s\n", playlist.Name)
// 	// }

// 	// for i := 0; i < 3; i++ {
// 	// 	resp, _ := p.Command(player.CMD_PLAYLISTS, 1, "foo", nil)

// 	// 	fmt.Printf("Resp: %v\n", resp)

// 	// time.Sleep(0 * time.Second)
// 	// }

// 	// TODO: Create audio.Decoder
// 	// TODO: Create audio.Output
// 	// TODO: Comment all.
// 	// TODO: Make initial commit.

// 	/*

// 	 SOME THOUGHTS.

// 	 * Decoder object includes TagReader.

// 	*/

// 	// tagReader := audio.NewTagReader(vfs.Test())
// 	// tag, err := tagReader.Parse(vfs.Test())
// 	// if err != nil {
// 	// 	fmt.Printf("Parsing error: %s\n", err)
// 	// }
// 	// fmt.Printf("Tag: %v\n\n", tag)

// 	// fs := vfs.New()
// 	// // err := fs.Append("foo")
// 	// // if err != nil {
// 	// // 	fmt.Println(err)
// 	// // }
// 	// fmt.Printf("%s\n", fs.Wd()) // /
// 	// fs.Cd("home")
// 	// fmt.Printf("%s\n", fs.Wd()) // /home
// 	// fs.Cd("viacheslav")
// 	// fmt.Printf("%s\n", fs.Wd()) // /home/viacheslav
// 	// fs.Cd("..")
// 	// fmt.Printf("%s\n", fs.Wd()) // /home
// 	// fs.Cd("..")
// 	// fs.Cd(".")
// 	// fs.Cd(".")
// 	// fmt.Printf("%s\n", fs.Wd()) // /
// 	// fs.Cd("viacheslav")
// 	// err := fs.Cd("/viacheslav/documents")
// 	// if err != nil {
// 	// 	fmt.Printf("Cd failed. %s\n", err)
// 	// } else {
// 	// 	fmt.Printf("%s\n", fs.Wd()) // /home/viacheslav/documents
// 	// }

// 	// fs.Cd("/")

// 	// err = fs.Cd("/viacheslav/.zshrc")
// 	// if err != nil {
// 	// 	fmt.Printf("Cd failed. %s\n", err)
// 	// } else {
// 	// 	fmt.Printf("%s\n", fs.Wd())
// 	// }
// }

func init() {
	// Register tag readers.
	vfs.RegisterTagReaderFactory(mp3.TagReaderFactory)
	vfs.RegisterTagReaderFactory(ogg.TagReaderFactory)

	// Register decoder drivers.
	audio.RegisterDecoderFactory(mp3.DecoderFactory)

	// Register audio output drivers.
	audio.RegisterOutputFactory(alsa.DriverName, alsa.OutputFactory)
}
