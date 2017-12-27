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

	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/serialize"
	"github.com/vchimishuk/chub/vfs"
)

type Client struct {
	conn    *ClientConn
	player  *player.Player
	close   chan interface{}
	onClose func(*Client, bool)
}

func newClient(conn net.Conn, pl *player.Player) *Client {
	return &Client{
		conn:   newClientConn(conn),
		player: pl,
		close:  make(chan interface{}, 1),
	}
}

func (c *Client) Serve() {
	c.conn.WriteLine("OK Chub v0.0")
	c.conn.WriteLine("")
	c.conn.Flush()

	kill := false
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
				kill = true
				quit = true
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

	c.conn.Close()
	c.close <- struct{}{}
	if c.onClose != nil {
		c.onClose(c, kill)
	}
}

func (c *Client) Close() {
	// Close connection to wake Server() up from blocking Read() or Write().
	c.conn.Close()
	<-c.close
}

func (c *Client) SetOnClose(handler func(*Client, bool)) {
	c.onClose = handler
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

func (c *Client) status() []string {
	s := c.player.State()

	if s == player.StateStopped {
		return []string{serialize.Map(map[string]interface{}{
			"state": "stopped",
		})}
	} else {
		var st string
		if s == player.StatePlaying {
			st = "playing"
		} else if s == player.StatePaused {
			st = "paused"
		} else {
			panic("invalid state")
		}

		p := c.player.CurPlaylist()
		t := c.player.Track()

		return []string{serialize.Map(map[string]interface{}{
			"state":             st,
			"playlist-name":     p.Name(),
			"playlist-length":   p.Len(),
			"playlist-position": c.player.PlaylistPos(),
			"track-path":        t.Path.String(),
			"track-position":    c.player.Pos(),
			"track-length":      t.Length,
		})}
	}
}
