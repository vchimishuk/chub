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

package proto

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"

	"github.com/vchimishuk/chub/serialize"
)

type Proto struct {
	conn   net.Conn
	reader *textproto.Reader
	writer *bufio.Writer
}

func New(conn net.Conn) *Proto {
	return &Proto{
		conn:   conn,
		reader: textproto.NewReader(bufio.NewReader(conn)),
		writer: bufio.NewWriter(conn),
	}
}

func (p *Proto) Close() error {
	return p.conn.Close()
}

func (p *Proto) WriteResponse(records []serialize.Serializable) error {
	s := make([]string, 0, len(records)+1)
	s = append(s, "OK")
	s = append(s, asStrings(records)...)

	return p.writeLines(s)
}

func (p *Proto) WriteError(e error) error {
	return p.writeLines([]string{fmt.Sprintf("ERR %s", e.Error())})
}

func (p *Proto) WriteEvent(name string,
	records []serialize.Serializable) error {

	s := make([]string, 0, len(records)+1)
	s = append(s, "EVENT "+name)
	s = append(s, asStrings(records)...)

	return p.writeLines(s)
}

func (p *Proto) ReadCommand() (*Command, error) {
	l, err := p.reader.ReadLine()
	if err != nil {
		return nil, err
	}

	return ParseCommand(l)
}

func (p *Proto) writeLines(lines []string) error {
	for _, line := range lines {
		err := p.writeLine(line)
		if err != nil {
			return err
		}
	}
	err := p.writeLine("")
	if err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Proto) writeLine(s string) error {
	_, err := p.writer.WriteString(s)
	if err != nil {
		return err
	}
	_, err = p.writer.WriteString("\n")

	return err
}

func asStrings(records []serialize.Serializable) []string {
	var s []string

	for _, r := range records {
		s = append(s, r.Serialize())
	}

	return s
}
