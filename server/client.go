// Copyright 2023 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package server

import (
	"errors"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/serialize"
	"github.com/vchimishuk/chub/server/proto"
	"github.com/vchimishuk/chub/vfs"
)

var errQuit = errors.New("normal quit")

type client struct {
	proto *proto.Proto
	// TODO: Can we remove Server dependency and use channel instead?
	srv      *Server
	player   *player.Player
	events   atomic.Bool
	writeMu  sync.Mutex
	close    sync.WaitGroup
	closeErr atomic.Value
}

func newClient(conn net.Conn, srv *Server, p *player.Player) *client {
	c := &client{
		proto:  proto.New(conn),
		srv:    srv,
		player: p,
	}
	c.close.Add(1)

	return c
}

func (c *client) Serve() {
	var err error
	var cmd *proto.Command

	for {
		var recs []serialize.Serializable
		cmd, err = c.proto.ReadCommand()
		if err != nil && !proto.IsError(err) {
			// Exit on network error.
			break
		}

		c.writeMu.Lock()

		if err == nil {
			switch cmd.Name {
			case proto.Events:
				c.events.Store(cmd.Args[0].(bool))
			case proto.Kill:
				err = errQuit
				// TODO: I guess I can find a better way to to it.
				go c.srv.Close()
			case proto.List:
				recs, err = c.list(cmd.Args[0].(string))
			case proto.Next:
				c.player.Next()
			case proto.Pause:
				c.player.Pause()
			case proto.Ping:
				// Do nothing.
			case proto.Play:
				err = c.play(cmd.Args[0].(string))
			case proto.PlaylistAppend:
				name := cmd.Args[0].(string)
				path := cmd.Args[1].(string)
				err = c.append(name, path)
			case proto.PlaylistClear:
				err = c.player.Clear(cmd.Args[0].(string))
			case proto.CreatePlaylist:
				err = c.player.Create(cmd.Args[0].(string))
			case proto.PlaylistDelete:
				err = c.player.Delete(cmd.Args[0].(string))
			case proto.PlaylistList:
				recs, err = c.playlist(cmd.Args[0].(string))
			case proto.PlaylistRename:
				oldName := cmd.Args[0].(string)
				newName := cmd.Args[1].(string)
				err = c.player.Rename(oldName, newName)
			case proto.Playlists:
				recs = c.playlists()
			case proto.Prev:
				c.player.Prev()
			case proto.Seek:
				c.player.Seek(cmd.Args[0].(int),
					cmd.Args[1].(bool))
			case proto.Status:
				recs = c.status()
			case proto.Stop:
				c.player.Stop()
			case proto.Quit:
				err = errQuit
			default:
				panic("unsupported command")
			}
		}

		if err == errQuit {
			// Exit on user request.
			c.writeMu.Unlock()
			break
		}

		var ioErr error
		if err != nil {
			// Do not show FS-sensitive information to user.
			if os.IsNotExist(err) || os.IsPermission(err) {
				err = errors.New("no such file or directory")
			}
			ioErr = c.proto.WriteError(err)
		} else {
			ioErr = c.proto.WriteResponse(recs)
		}
		c.writeMu.Unlock()

		if ioErr != nil {
			// Exit on network error.
			err = ioErr
			break
		}
	}

	c.proto.Close()
	c.closeErr.Store(err)
	c.close.Done()
}

func (c *client) Close() error {
	// Close connection to wake Serve() up from blocking Read() or Write().
	c.proto.Close()
	c.close.Wait()

	return c.closeErr.Load().(error)
}

func (c *client) IsClosed() bool {
	return c.closeErr.Load() != nil
}

func (c *client) Notify(e player.Event) error {
	if !c.events.Load() {
		return nil
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.proto.WriteEvent(e.Name(), e.Body())
}

func (c *client) play(path string) error {
	p, err := vfs.NewPath(path)
	if err != nil {
		return err
	}

	return c.player.Play(p)
}

func (c *client) append(name string, path string) error {
	p, err := vfs.NewPath(path)
	if err != nil {
		return err
	}

	return c.player.Append(name, p)
}

func (c *client) list(path string) ([]serialize.Serializable, error) {
	p, err := vfs.NewPath(path)
	if err != nil {
		return nil, err
	}
	if !p.IsDir() {
		return nil, errors.New("not a directory")
	}
	es, err := p.List()

	return serializableSlice(es), err
}

func (c *client) playlist(name string) ([]serialize.Serializable, error) {
	plist, err := c.player.Playlist(name)
	if err != nil {
		return nil, err
	}

	return serializableSlice(plist.Tracks()), nil
}

func (c *client) playlists() []serialize.Serializable {

	return serializableSlice(c.player.Playlists())
}

// TODO: Return maps and let the caller to serialize it?
func (c *client) status() []serialize.Serializable {
	st := c.player.Status()

	if st.State == player.StateStopped {
		return []serialize.Serializable{
			serialize.Wrap(map[string]interface{}{
				"state": "stopped",
			}),
		}
	} else {
		var s string
		if st.State == player.StatePlaying {
			s = "playing"
		} else if st.State == player.StatePaused {
			s = "paused"
		} else {
			panic("invalid state")
		}
		track := st.Plist.Get(st.PlistPos)

		return []serialize.Serializable{
			serialize.Wrap(map[string]interface{}{
				"state":             s,
				"playlist-position": st.PlistPos,
				"track-position":    st.Pos,
				"playlist-name":     st.Plist.Name(),
				"playlist-duration": st.Plist.Duration(),
				"playlist-length":   st.Plist.Len(),
				"track-path":        track.Path.String(),
				"track-artist":      track.Tag.Artist,
				"track-album":       track.Tag.Album,
				"track-title":       track.Tag.Title,
				"track-number":      track.Tag.Number,
				"track-length":      track.Length,
			}),
		}
	}
}

func serializableSlice[E serialize.Serializable](s []E) []serialize.Serializable {
	t := make([]serialize.Serializable, 0, len(s))
	for _, e := range s {
		t = append(t, e)
	}

	return t
}
