package server

import (
	"../player"
	"../vfs"
	"bufio"
)

// Client handler object.
type clientHandler struct {
	player *player.Player
	out    *bufio.Writer
}

// newClientHandler creates new handler for new client connection.
func newClientHandler(pl *player.Player, out *bufio.Writer) *clientHandler {
	return &clientHandler{player: pl, out: out}
}

// Init initializes client communication protocol.
// Sends greeting to the client.
func (cl *clientHandler) Init() {
	// TODO: Move version to some more appropriate place.
	cl.ok().writeLn("Chub 0.0").eom()
}

// HandleCommand handles one client command and writes result to the client.
func (cl *clientHandler) HandleCommand(cmd *command) bool {
	var quit bool = false

	switch cmd.name {
	case cmdAdd:
		name := cmd.args[0].(string)
		path := cmd.args[1].(string)
		cl.cmdAdd(name, path)
	case cmdAddPlaylist:
		name := cmd.args[0].(string)
		cl.cmdAddPlaylist(name)
	case cmdLs:
		dir := cmd.args[0].(string)
		cl.cmdLs(dir)
	case cmdPing:
		cl.cmdPing()
	case cmdPlaylists:
		cl.cmdPlaylists()
	case cmdPlaylist:
		name := cmd.args[0].(string)
		cl.cmdPlaylist(name)
	case cmdQuit:
		cl.cmdQuit()
		quit = true
	default:
		panic("Unsupported command received.")
	}

	cl.eom()

	return quit
}

// ADD command handler.
func (cl *clientHandler) cmdAdd(name string, path string) {
	err := cl.player.Add(name, vfs.NewPath(path))
	if err != nil {
		cl.WriteError(err.Error())
	} else {
		cl.ok()
	}
}

// ADDPLAYLIST command handler.
func (cl *clientHandler) cmdAddPlaylist(name string) {
	err := cl.player.AddPlaylist(name)
	if err != nil {
		cl.WriteError(err.Error())
	} else {
		cl.ok()
	}
}

// LS command handler.
func (cl *clientHandler) cmdLs(dir string) {
	path := vfs.NewPath(dir)

	entries, err := path.List()
	if err != nil {
		cl.WriteError(err.Error())
	} else {
		cl.ok()
		for _, e := range entries {
			switch e.(type) {
			case *vfs.Track:
				track := e.(*vfs.Track)
				tag := track.Tag
				m := map[string]interface{}{
					"Type":   "TRACK",
					"Artist": tag.Artist,
					"Album":  tag.Album,
					"Title":  tag.Title,
					"Number": tag.Number,
					"Length": tag.Length,
				}
				writeMap(cl.out, m)
				cl.writeLn("")
			case *vfs.Directory:
				d := e.(*vfs.Directory)
				m := map[string]interface{}{
					"Type": "DIRECTORY",
					"Name": d.Name,
					"Path": d.Path.String(),
				}
				writeMap(cl.out, m)
				cl.writeLn("")
			default:
				panic("Unsupported entry type.")
			}
		}
	}
}

// PING client command handler.
func (cl *clientHandler) cmdPing() {
	cl.ok()
}

// PLAYLISTS client command handler.
func (cl *clientHandler) cmdPlaylists() {
	cl.ok()

	for _, plist := range cl.player.Playlists() {
		m := map[string]interface{}{
			"Name":   plist.Name(),
			"Length": plist.Len(),
		}
		writeMap(cl.out, m)
		cl.writeLn("")
	}
}

// PLAYLIST client command handler.
func (cl *clientHandler) cmdPlaylist(name string) {
	plist, err := cl.player.Playlist(name)

	if err != nil {
		cl.WriteError(err.Error())
	} else {
		cl.ok()

		for i := 0; i < plist.Len(); i++ {
			track := plist.Get(i)
			m := map[string]interface{}{
				"Path":   track.Path.String(),
				"Artist": track.Tag.Artist,
				"Album":  track.Tag.Album,
				"Title":  track.Tag.Title,
			}
			writeMap(cl.out, m)
			cl.writeLn("")
		}
	}
}

// QUIT client command handler.
func (cl *clientHandler) cmdQuit() {
	cl.ok().writeLn("Bye.")
}

// WriteError writes an error message to the client.
func (cl *clientHandler) WriteError(err string) {
	cl.error().writeLn(err).eom()
}

// TODO: Log all output with debug.

// Writes error response message header.
func (cl *clientHandler) error() *clientHandler {
	cl.writeLn("ERR")

	return cl
}

// Writes regular (successful) response message header.
func (cl *clientHandler) ok() *clientHandler {
	cl.writeLn("OK")

	return cl
}

// Writes response message footer (end of message marker).
func (cl *clientHandler) eom() {
	cl.writeLn("")
	cl.out.Flush()
}

// Writes string to the client.
func (cl *clientHandler) write(s string) *clientHandler {
	cl.out.WriteString(s)

	return cl
}

// Sends '\n' terminated string to the client.
func (cl *clientHandler) writeLn(s string) *clientHandler {
	cl.out.WriteString(s)
	cl.out.WriteString("\n")

	return cl
}
