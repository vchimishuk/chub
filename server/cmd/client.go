package cmd

import (
	"errors"
	"net"
	"os"

	"github.com/vchimishuk/chub/player"
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
		close:  make(chan interface{}, 1)}
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
			var resp []map[string]interface{}

			switch cmd.name {
			case cmdKill:
				kill = true
				quit = true
			case cmdLs:
				resp, err = c.ls(cmd.args[0].(string))
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
			case cmdPlaylistCreate:
				err = c.player.Create(cmd.args[0].(string))
			case cmdPlaylistDelete:
				err = c.player.Delete(cmd.args[0].(string))
			case cmdPlaylistList:
				resp, err = c.playlist(cmd.args[0].(string))
			case cmdPlaylistRename:
				oldName := cmd.args[0].(string)
				newName := cmd.args[1].(string)
				err = c.player.Rename(oldName, newName)
			case cmdPlaylistsList:
				resp = c.playlists()
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
				c.conn.WriteOkResp(resp...)
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

func (c *Client) OnClose(handler func(*Client, bool)) {
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

func (c *Client) ls(path string) ([]map[string]interface{}, error) {
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

	maps := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		maps = append(maps, entryToMap(e))
	}

	return maps, nil
}

func (c *Client) playlist(name string) ([]map[string]interface{}, error) {
	plist, err := c.player.Playlist(name)
	if err != nil {
		return nil, err
	}

	maps := make([]map[string]interface{}, 0, len(plist.Tracks()))
	for _, track := range plist.Tracks() {
		maps = append(maps, trackToMap(track))
	}

	return maps, nil
}

func (c *Client) playlists() []map[string]interface{} {
	plists := c.player.Playlists()
	maps := make([]map[string]interface{}, 0, len(plists))
	for _, pl := range plists {
		maps = append(maps, map[string]interface{}{
			"name":     pl.Name,
			"duration": pl.Duration,
			"length":   pl.Len,
		})
	}

	return maps
}
