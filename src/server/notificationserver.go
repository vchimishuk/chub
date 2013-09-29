package server

import (
	"../logger"
	"bufio"
	"net"
	"strconv"
	"sync"
)

// List of notification events can be sent to client.
const (
	ntfPlaylistChanged  = "PLAYLIST_CHANGED"
	ntfPlaylistsChanged = "PLAYLISTS_CHANGED"
	ntfStateChanged     = "STATE_CHANGED"
	ntfTrackChanged     = "TRACK_CHANGED"
	ntfVolumeChanged    = "VOLUME_CHANGED"
)

// NotificationServer serves client connections for notification protocol.
// Every client connected to this server receives notification about any
// player state change events: current track was changed, player was paused,
// and so on.
type NotificationServer struct {
	// Server listener.
	listener *net.TCPListener
	// All connected client connections which will be notified.
	connections []net.Conn
	// Mutex to syncronize connections array access.
	connLock sync.Mutex
}

// NewNotificationServer creates new server.
func NewNotificationServer() *NotificationServer {
	return new(NotificationServer)
}

// Serve starts listening and handling incoming client connections.
func (srv *NotificationServer) Serve(addr string, port int) error {
	ip, err := resolveAddr(addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: ip, Port: port})
	if err != nil {
		return err
	}
	logger.Info("NotificationServer started listening %s:%d", ip, port)

	srv.listener = listener

	go func() {
		for {
			conn, err := srv.listener.AcceptTCP()
			if err != nil {
				logger.Error("Accepting client failed. Exiting.")
				return
			}

			logger.Info("NotificationServer: Client connected.")

			srv.connLock.Lock()
			srv.connections = append(srv.connections, conn)
			srv.connLock.Unlock()
		}
	}()

	return nil
}

func (srv *NotificationServer) Stop() {
	// TODO
}

// StateListener interface implementation.
func (srv *NotificationServer) PlaylistChanged(name string) {
	srv.notify(ntfPlaylistChanged, name)
}

// StateListener interface implementation.
func (srv *NotificationServer) PlaylistsChanged() {
	srv.notify(ntfPlaylistsChanged)
}

// StateListener interface implementation.
func (srv *NotificationServer) StateChanged() {

}

// StateListener interface implementation.
func (srv *NotificationServer) TrackChanged() {

}

// StateListener interface implementation.
func (srv *NotificationServer) VolumeChanged() {

}

// Send notification to all clients.
func (srv *NotificationServer) notify(event string, params ...interface{}) {
	go func() {
		srv.connLock.Lock()
		defer srv.connLock.Unlock()

		for i := len(srv.connections) - 1; i >= 0; i-- {
			conn := srv.connections[i]
			err := srv.write(conn, event, params...)
			if err != nil {
				conn.Close()
				// Remove connection.
				copy(srv.connections[i:], srv.connections[i+1:])
			}
		}
	}()
}

// Write notification name and parameters to the client connection.
func (srv *NotificationServer) write(conn net.Conn, event string, params ...interface{}) error {
	out := bufio.NewWriter(conn)

	_, err := out.WriteString(event)
	if err != nil {
		// Connection was closed.
		return err
	}

	for i, p := range params {
		var s string

		if i == 0 {
			s = " "
		} else {
			s = ", "
		}

		switch p.(type) {
		case int:
			s += strconv.Itoa(p.(int))
		case string:
			s += strconv.Quote(p.(string))
		default:
			panic("Unsupported parameter type.")
		}

		if _, err := out.WriteString(s); err != nil {
			return err
		}
	}

	_, err = out.WriteString("\n")
	if err != nil {
		return err
	}
	out.Flush()

	return nil
}
