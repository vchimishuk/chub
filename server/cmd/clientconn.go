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
