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

	"github.com/vchimishuk/chub/cnet"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/serialize"
)

type Client struct {
	conn    *cnet.TextConn
	notif   chan *player.NotifMsg
	close   chan struct{}
	onClose func(*Client, bool)
}

func newClient(conn net.Conn, notif chan *player.NotifMsg) *Client {
	return &Client{
		conn:  cnet.NewTextConn(conn),
		notif: notif,
		close: make(chan struct{}, 1),
	}
}

func (c *Client) Serve() {
	for {
		select {
		case msg := <-c.notif:
			c.notify(msg)
		case <-c.close:
			break
		}
	}

	c.conn.Close()
}

func (c *Client) Close() {
	c.close <- struct{}{}
	<-c.close
}

func (c *Client) SetOnClose(handler func(*Client, bool)) {
	c.onClose = handler
}

func (c *Client) notify(msg *player.NotifMsg) {
	c.conn.WriteLine(string(msg.Event))

	switch msg.Event {
	case player.PlaylistsEvent:
		plists := msg.Value.([]*player.PlaylistInfo)
		c.playlists(plists)
	default:
		panic("unsupported event")
	}

	c.conn.WriteLine("")
	c.conn.Flush()
}

func (c *Client) playlists(plists []*player.PlaylistInfo) {
	for _, pl := range plists {
		c.conn.WriteLine(serialize.PlaylistInfo(pl))
	}
}
