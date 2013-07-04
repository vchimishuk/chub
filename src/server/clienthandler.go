package server

import (
	"../player"
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
	cl.ok()
	cl.writeString("Chub 0.0")
	cl.eom()
}

// HandleCommand handles one client command and writes result to the client.
func (cl *clientHandler) HandleCommand(cmd *command) bool {
	var quit bool = false

	switch cmd.name {
	case cmdPing:
		cl.cmdPing()
	case cmdQuit:
		cl.cmdQuit()
		quit = true
	default:
		panic("Unsupported command received.")
	}

	cl.eom()

	return quit
}

// PING client command handler.
func (cl *clientHandler) cmdPing() {
	cl.ok()
}

// QUIT client command handler.
func (cl *clientHandler) cmdQuit() {
	cl.ok()
	cl.writeString("Bye.")
}

// WriteError writes an error message to the client.
func (cl *clientHandler) WriteError(err string) {
	cl.error()
	cl.writeString(err)
	cl.eom()
}

// TODO: Log all output with debug.

// Writes error response message header.
func (cl *clientHandler) error() {
	cl.writeString("ERR")
}

// Writes regular (successful) response message header.
func (cl *clientHandler) ok() {
	cl.writeString("OK")
}

// Writes response message footer (end of message marker).
func (cl *clientHandler) eom() {
	cl.writeString("")
	cl.out.Flush()
}

// Sends '\n' terminated string to the client.
func (cl *clientHandler) writeString(s string) {
	cl.out.WriteString(s)
	cl.out.WriteString("\n")
}
