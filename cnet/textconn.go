package cnet

import (
	"bufio"
	"net"
	"net/textproto"
)

type TextConn struct {
	conn   net.Conn
	reader *textproto.Reader
	writer *bufio.Writer
}

func NewTextConn(conn net.Conn) *TextConn {
	return &TextConn{
		conn:   conn,
		reader: textproto.NewReader(bufio.NewReader(conn)),
		writer: bufio.NewWriter(conn)}
}

func (c *TextConn) ReadLine() (string, error) {
	return c.reader.ReadLine()
}

func (c *TextConn) WriteLine(line string) (int, error) {
	n, err := c.writer.WriteString(line)
	if err != nil {
		return n, err
	}
	n, err = c.writer.WriteString("\n")

	return n, err
}

func (c *TextConn) Flush() error {
	return c.writer.Flush()
}

func (c *TextConn) Close() error {
	return c.conn.Close()
}
