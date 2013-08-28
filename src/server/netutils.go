package server

import (
	"fmt"
	"net"
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
