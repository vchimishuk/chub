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
	"fmt"
	"net"

	"github.com/vchimishuk/chub/cnet"
)

type CmdConn struct {
	*cnet.TextConn
}

func newCmdConn(conn net.Conn) *CmdConn {
	return &CmdConn{TextConn: cnet.NewTextConn(conn)}
}

func (c *CmdConn) WriteOkResp(lines []string) error {
	_, err := c.WriteLine("OK")

	for _, line := range lines {
		_, err := c.WriteLine(line)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	_, err = c.WriteLine("")

	return err
}

func (c *CmdConn) WriteErrorResp(e error) error {
	_, err := c.WriteLine(fmt.Sprintf("ERR %s", e.Error()))
	if err != nil {
		return err
	}
	_, err = c.WriteLine("")

	return err
}
