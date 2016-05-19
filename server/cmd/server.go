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

package cmd

import (
	"net"

	"github.com/vchimishuk/chub/cnet"
	"github.com/vchimishuk/chub/player"
)

type Server struct {
	srv *cnet.Server
}

func NewServer(p *player.Player) *Server {
	srv := cnet.NewServer()
	srv.SetOnClient(func(conn net.Conn) cnet.Client {
		c := newClient(conn, p)
		c.SetOnClose(func(cl *Client, kill bool) {
			srv.RemoveClient(cl)
			if kill {
				srv.Close()
			}
		})
		go c.Serve()

		return c
	})

	return &Server{srv: srv}
}

func (s *Server) Listen(addr string, port int) error {
	return s.srv.Listen(addr, port)
}

func (s *Server) Serve() {
	s.srv.Serve()
}

func (s *Server) Close() {
	s.srv.Close()
}
