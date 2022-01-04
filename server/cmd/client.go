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
	"errors"
	"net"
	"os"
	"sync"

	"github.com/vchimishuk/chub/cnet"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/serialize"
	"github.com/vchimishuk/chub/vfs"
)

type Client struct {
	conn     *CmdConn
	srv      *cnet.Server
	player   *player.Player
	close    chan interface{}
	closedMu sync.Mutex
	closed   bool
}

func NewClient(conn net.Conn, srv *cnet.Server, p *player.Player) *Client {
	return &Client{
		conn:   newCmdConn(conn),
		srv:    srv,
		player: p,
		close:  make(chan interface{}, 1),
	}
}

func (c *Client) Serve() {
	c.conn.WriteLine("OK Chub v0.0")
	c.conn.WriteLine("")
	c.conn.Flush()

	quit := false

	for !quit {
		line, err := c.conn.ReadLine()
		if err != nil {
			// TODO: Log.
			break
		}

		cmd, err := parseCommand(line)
		if err != nil {
			c.conn.WriteErrorResp(err)
		} else {
			var err error
			var lines []string

			switch cmd.name {
			case cmdKill:
				quit = true
				go c.srv.Close()
			case cmdList:
				lines, err = c.list(cmd.args[0].(string))
			case cmdNext:
				c.player.Next()
			case cmdPause:
				c.player.Pause()
			case cmdPing:
				// Do nothing.
			case cmdPlay:
				err = c.play(cmd.args[0].(string))
			case cmdPlaylistAppend:
				name := cmd.args[0].(string)
				path := cmd.args[1].(string)
				err = c.append(name, path)
			case cmdPlaylistClear:
				err = c.player.Clear(cmd.args[0].(string))
			case cmdCreatePlaylist:
				err = c.player.Create(cmd.args[0].(string))
			case cmdPlaylistDelete:
				err = c.player.Delete(cmd.args[0].(string))
			case cmdPlaylistList:
				lines, err = c.playlist(cmd.args[0].(string))
			case cmdPlaylistRename:
				oldName := cmd.args[0].(string)
				newName := cmd.args[1].(string)
				err = c.player.Rename(oldName, newName)
			case cmdPlaylists:
				lines = c.playlists()
			case cmdPrev:
				c.player.Prev()
			case cmdStatus:
				lines = c.status()
			case cmdStop:
				c.player.Stop()
			case cmdQuit:
				quit = true
			default:
				err := errors.New("unsupported command")
				c.conn.WriteErrorResp(err)
			}

			if err != nil {
				if err != nil && (os.IsNotExist(err) ||
					os.IsPermission(err)) {
					err = errors.New("no such file or directory")
				}
				c.conn.WriteErrorResp(err)
			} else {
				c.conn.WriteOkResp(lines)
			}
		}

		err = c.conn.Flush()
		if err != nil {
			// TODO: Log.
			break
		}
	}

	c.closedMu.Lock()
	c.closed = true
	c.closedMu.Unlock()
	c.conn.Close()
	c.close <- struct{}{}
}

func (c *Client) Close() error {
	// Close connection to wake Server() up from blocking Read() or Write().
	err := c.conn.Close()
	<-c.close
	c.closedMu.Lock()
	c.closed = true
	c.closedMu.Unlock()

	return err
}

func (c *Client) IsClosed() bool {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	return c.closed
}

func (c *Client) play(path string) error {
	p, err := vfs.NewPath(path)
	if err != nil {
		return err
	}

	return c.player.Play(p)
}

func (c *Client) append(name string, path string) error {
	p, err := vfs.NewPath(path)
	if err != nil {
		return err
	}

	return c.player.Append(name, p)
}

func (c *Client) list(path string) ([]string, error) {
	p, err := vfs.NewPath(path)
	if err != nil {
		return nil, err
	}
	if !p.IsDir() {
		return nil, errors.New("not a directory")
	}
	entries, err := p.List()
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, serialize.Entry(e))
	}

	return lines, nil
}

func (c *Client) playlist(name string) ([]string, error) {
	plist, err := c.player.Playlist(name)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0, plist.Len())
	for i := 0; i < plist.Len(); i++ {
		lines = append(lines, serialize.Track(plist.Get(i)))
	}

	return lines, nil
}

func (c *Client) playlists() []string {
	plists := c.player.Playlists()
	lines := make([]string, 0, len(plists))
	for _, pl := range plists {
		lines = append(lines, serialize.Playlist(pl))
	}

	return lines
}

// TODO: Return maps and let the caller to serialize it.
func (c *Client) status() []string {
	st := c.player.Status()

	//	s := c.player.State()

	if st.State == player.StateStopped {
		return []string{serialize.Map(map[string]interface{}{
			"state": "stopped",
		})}
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

		return []string{serialize.Map(map[string]interface{}{
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
		})}
	}
}
