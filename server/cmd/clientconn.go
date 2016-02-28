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
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/vchimishuk/chub/cnet"
)

type ClientConn struct {
	*cnet.TextConn
}

func newClientConn(conn net.Conn) *ClientConn {
	return &ClientConn{TextConn: cnet.NewTextConn(conn)}
}

func (c *ClientConn) WriteOkResp(items ...map[string]interface{}) error {
	_, err := c.WriteLine("OK")

	for _, item := range items {
		_, err := c.WriteLine(mapToString(item))
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

func (c *ClientConn) WriteErrorResp(e error) error {
	_, err := c.WriteLine(fmt.Sprintf("ERR %s", e.Error()))
	if err != nil {
		return err
	}
	_, err = c.WriteLine("")

	return err
}

func mapToString(m map[string]interface{}) string {
	var b bytes.Buffer
	var l int = len(m)
	var i int = 1

	for k, v := range m {
		b.WriteString(k)
		b.WriteString(": ")

		switch v.(type) {
		case int:
			b.WriteString(strconv.Itoa(v.(int)))
		case string:
			b.WriteString(strconv.Quote(v.(string)))
		case bool:
			b.WriteString(strconv.FormatBool(v.(bool)))
		default:
			panic("Unsupported type")
		}

		if i < l {
			b.WriteString(", ")
		}
		i++
	}

	return b.String()
}
