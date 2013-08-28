// server package incapsulates code which handle client connections and
// map client commands to the core player engine.
package server

import (
	"../logger"
	"../player"
	"bufio"
	"fmt"
	"net"
	"net/textproto"
)

// CommandServer object.
type CommandServer struct {
	player   *player.Player
	listener *net.TCPListener
}

// New returns newly initialized CommandServer object.
func NewCommandServer(pl *player.Player) *CommandServer {
	return &CommandServer{player: pl}
}

// Serve starts listening and handling incoming client connections.
func (srv *CommandServer) Serve(addr string, port int) error {
	ip, err := resolveAddr(addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: ip, Port: port})
	if err != nil {
		return err
	}
	logger.Info("CommandServer started listening %s:%d", ip, port)

	srv.listener = listener

	go func() {
		for {
			conn, err := srv.listener.AcceptTCP()
			if err != nil {
				logger.Error("Accepting client failed. Exiting.")
				return
			}

			go srv.handler(conn)
		}
	}()

	return nil
}

// Stop stops client connections handling and shutdown the server.
func (srv *CommandServer) Stop() {
	// TODO:
}

// handle handle one particular client connection.
func (srv *CommandServer) handler(conn *net.TCPConn) {
	// This function is not threadsafe and should not modify srv object.
	cid := conn // Client's uniq id.
	writer := bufio.NewWriter(conn)
	reader := textproto.NewReader(bufio.NewReader(conn))

	defer conn.Close()
	defer logger.Info("Client %d disconnected.", cid)

	logger.Info("New client %d accepted. Addr: %s", cid, conn.RemoteAddr())

	client := newClientHandler(srv.player, writer)
	client.Init()

	for {
		line, err := reader.ReadLine()
		if err != nil {
			// Connection was closed or something like that.
			break
		}

		logger.Info("Client %d command received: '%s'.", cid, line)

		cmd, err := parseCommand(line)
		if err != nil {
			client.WriteError(fmt.Sprintf("Failed to parse command. %s", err))
		} else {
			quit := client.HandleCommand(cmd)
			if quit {
				break
			}
		}
	}
}
