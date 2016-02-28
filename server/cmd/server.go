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
// along with Chub.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/vchimishuk/chub/player"
)

type Server struct {
	player    *player.Player
	listener  *net.TCPListener
	close     chan interface{}
	clientsMu sync.Mutex // Guards following field.
	clients   []*Client
}

func New(p *player.Player) *Server {
	return &Server{player: p, close: make(chan interface{}, 1)}
}

func (s *Server) Listen(addr string, port int) error {
	ip, err := resolveAddr(addr)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: ip, Port: port})
	if err != nil {
		return err
	}
	s.listener = listener

	return nil
}

func (s *Server) Serve() {
	// Wait berore retry on errors.
	const maxDelay = time.Second
	var delay time.Duration

	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				if delay == 0 {
					delay = 50 * time.Millisecond
				} else {
					delay *= 2
				}
				if delay > maxDelay {
					delay = maxDelay
				}
				time.Sleep(delay)
			} else {
				break
			}
		} else {
			delay = 0

			c := newClient(conn, s.player)
			c.OnClose(func(c *Client, kill bool) {
				s.removeClient(c)
				if kill {
					s.listener.Close()
				}
			})
			s.addClient(c)
			go c.Serve()
		}
	}

	s.player.Stop()
	s.closeClients()
	s.close <- struct{}{}
}

func (s *Server) Close() {
	// Close listener to wake the server up from AcceptTCP().
	s.listener.Close()
	<-s.close
}

func (s *Server) addClient(c *Client) {
	s.clientsMu.Lock()
	s.clients = append(s.clients, c)
	s.clientsMu.Unlock()
}

func (s *Server) removeClient(c *Client) {
	s.clientsMu.Lock()
	for i, cc := range s.clients {
		if cc == c {
			s.clients = append(s.clients[:i],
				s.clients[i+1:]...)
			break
		}
	}
	s.clientsMu.Unlock()
}

func (s *Server) closeClients() {
	s.clientsMu.Lock()
	for _, c := range s.clients {
		c.Close()
	}
	s.clients = make([]*Client, 0)
	s.clientsMu.Unlock()
}

func resolveAddr(addr string) (ip net.IP, err error) {
	ip = net.ParseIP(addr)
	if ip == nil {
		ips, err := net.LookupIP(addr)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			err = fmt.Errorf("Unable to resolve %s hostname.", addr)

			return nil, err
		}

		ip = ips[0]
	}

	return ip, nil
}
