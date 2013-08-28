package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
)

// resolveAddr tries to handle addr as IP first and if it can't be parsed
// hanldes it as a host name.
func resolveAddr(addr string) (ip net.IP, err error) {
	ip = net.ParseIP(addr)
	if ip == nil {
		ips, err := net.LookupIP(addr)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			err = fmt.Errorf("Failed to resolve %s hostname. %s", addr)

			return nil, err
		}

		ip = ips[0]
	}

	return ip, nil
}

// writeMap serializes map to JSON-like string and write output to the
// given writer. For instance, map {"foo": "bar", "baz": 123} will be serialized
// to string "foo: "bar", baz: 123".
func writeMap(w io.Writer, m map[string]interface{}) (n int, err error) {
	first := true
	n = 0
	writer := bufio.NewWriter(w)

	for k, v := range m {
		var nn int

		if first {
			first = false
		} else {
			if nn, err = writer.WriteString(", "); err != nil {
				return n + nn, err
			}
			n += nn
		}

		if nn, err = writer.WriteString(k); err != nil {
			return n + nn, err
		}
		n += nn
		if nn, err = writer.WriteString(": "); err != nil {
			return n + nn, err
		}
		n += nn

		var s string

		switch v.(type) {
		case int:
			s = strconv.Itoa(v.(int))
		case string:
			s = strconv.Quote(v.(string))
		default:
			panic("Unsupported map value.")
		}

		if nn, err = writer.WriteString(s); err != nil {
			return n + nn, err
		}
		n += nn
	}

	return n, nil
}
