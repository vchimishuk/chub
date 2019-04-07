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

package notif

import (
	"net"
	"sync"

	"github.com/vchimishuk/chub/cnet"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/serialize"
)

type responseLine map[string]interface{}

type Client struct {
	conn     *cnet.TextConn
	closedMu sync.Mutex
	closed   bool
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: cnet.NewTextConn(conn)}
}

func (c *Client) Close() error {
	c.closedMu.Lock()
	err := c.conn.Close()
	c.closed = true
	c.closedMu.Unlock()

	return err
}

func (c *Client) Serve() {

}

func (c *Client) IsClosed() bool {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	return c.closed
}

func (c *Client) Notify(e player.Event, args []interface{}) error {
	_, err := c.conn.WriteLine(string(e))
	if err != nil {
		c.Close()
		return err
	}

	var lines []responseLine

	switch e {
	case player.EventStatus:
		lines = c.status(args[0].(*player.Status))
	// case player.PlaylistEvent:
	// 	name := msg.Args[0].(string)
	// 	tracks := msg.Args[1].([]*vfs.Track)
	// 	c.playlist(name, tracks)
	// case player.PlaylistsEvent:
	// 	plists := msg.Args[0].([]*player.Playlist)
	// 	c.playlists(plists)
	default:
		panic("unsupported event")
	}

	for _, l := range lines {
		_, err := c.conn.WriteLine(serialize.Map(l))
		if err != nil {
			c.Close()
			return err
		}

	}
	_, err = c.conn.WriteLine("")
	if err != nil {
		c.Close()
		return err
	}
	err = c.conn.Flush()
	if err != nil {
		c.Close()
		return err
	}

	return nil
}

// func (c *Client) playlists(plists []*player.Playlist) {
// 	for _, pl := range plists {
// 		c.conn.WriteLine(serialize.Playlist(pl))
// 	}
// }

// func (c *Client) playlist(name string, tracks []*vfs.Track) {
// 	c.conn.WriteLine(name)
// 	for _, t := range tracks {
// 		c.conn.WriteLine(serialize.Track(t))
// 	}
// }

func (c *Client) status(st *player.Status) []responseLine {
	if st.State == player.StateStopped {
		return []responseLine{{"state": "stopped"}}
	} else {
		s := ""
		if st.State == player.StatePlaying {
			s = "playing"
		} else if st.State == player.StatePaused {
			s = "paused"
		} else {
			panic("invalid state")
		}

		track := st.Plist.Get(st.PlistPos)

		return []responseLine{
			{"state": s},
			{"playlist-position": st.PlistPos},
			{"track-position": st.Pos},
			{"playlist-name": st.Plist.Name()},
			{"playlist-duration": st.Plist.Duration()},
			{"playlist-length": st.Plist.Len()},
			{"track-path": track.Path.String()},
			{"track-artist": track.Tag.Artist},
			{"track-album": track.Tag.Album},
			{"track-title": track.Tag.Title},
			{"track-number": track.Tag.Number},
			{"track-length": track.Length},
		}
	}
}
