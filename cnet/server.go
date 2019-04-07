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

package cnet

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Client interface {
	Serve()
	Close() error
	IsClosed() bool
}

type ClientHandler func(conn net.Conn, srv *Server) Client

type Server struct {
	listener *net.TCPListener
	close    chan struct{}
	clients  sync.Map
	handler  ClientHandler
}

func NewServer(h ClientHandler) *Server {
	return &Server{
		close:   make(chan struct{}),
		handler: h,
	}
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
		} else {
			s.cleanUpClients()
			delay = 0
			c := s.handler(conn, s)
			s.clients.Store(c, struct{}{})
			go c.Serve()
		}
	}

	s.clients.Range(func(k, v interface{}) bool {
		k.(Client).Close()
		s.clients.Delete(k)

		return true
	})
	close(s.close)
}

func (s *Server) Close() {
	// Close listener to wake the server up from AcceptTCP().
	s.listener.Close()
	<-s.close
}

func (s *Server) Clients() []Client {
	cs := []Client{}
	s.cleanUpClients()
	s.clients.Range(func(k, v interface{}) bool {
		cs = append(cs, k.(Client))
		return true
	})

	return cs
}

func (s *Server) cleanUpClients() {
	s.clients.Range(func(k, v interface{}) bool {
		c := k.(Client)
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
