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
