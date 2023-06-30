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
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/vchimishuk/chub/csync/job"
	"github.com/vchimishuk/chub/player"
)

type Server struct {
	player        *player.Player
	listener      *net.TCPListener
	listenerClose chan struct{}
	notifier      job.Job
	clients       sync.Map
}

func New(p *player.Player) *Server {
	s := &Server{
		player:        p,
		listenerClose: make(chan struct{}),
	}
	s.notifier = job.Start(s.notify)

	return s
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
				// TODO: Return error.
				break
			}

			continue
		}

		s.cleanUpClients()
		delay = 0
		c := newClient(conn, s, s.player)
		s.clients.Store(c, struct{}{})
		go c.Serve()
	}

	s.clients.Range(func(k, v interface{}) bool {
		c := k.(*client)
		c.Close()
		s.clients.Delete(k)

		return true
	})
}

func (s *Server) Close() {
	s.notifier.Shutdown()

	// Close listener to wake the server up from AcceptTCP().
	s.listener.Close()
	// And wait for server shutdown.
	<-s.listenerClose
}

func (s *Server) notify(close <-chan any) error {
	events := s.player.Events()

loop:
	for {
		select {
		case e := <-events:
			if e == nil {
				break
			}
			s.clients.Range(func(k, v interface{}) bool {
				c := k.(*client)
				go func() {
					err := c.Notify(e)
					if err != nil {
						c.Close()
					}
				}()
				return true
			})
		case <-close:
			break loop
		}
	}

	return nil
}

func (s *Server) cleanUpClients() {
	s.clients.Range(func(k, v interface{}) bool {
		c := k.(*client)
		if c.IsClosed() {
			s.clients.Delete(k)
		}

		return true
	})
}

func resolveAddr(addr string) (ip net.IP, err error) {
	ip = net.ParseIP(addr)
	if ip == nil {
		ips, err := net.LookupIP(addr)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			err = fmt.Errorf("unable to resolve %s hostname", addr)

			return nil, err
		}

		ip = ips[0]
	}

	return ip, nil
}
