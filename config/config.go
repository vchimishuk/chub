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

// Confiruration files parser implementation.
//
// Configuration file expected to be a traditional UNIX
// key-value plain text file. E.g.
//
// foo.bar = "some value"
// foo.baz = "some another value"
// bar = 14
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Error struct {
	Line    int
	Message string
}

func (err Error) Error() string {
	return fmt.Sprintf("%d: %s", err.Line, err.Message)
}

type Config struct {
	data map[string]string
}

func (c *Config) Defined(name string) bool {
	_, ok := c.data[name]

	return ok
}

func (c *Config) String(name string, def string) string {
	if val, ok := c.data[name]; ok {
		return val
	} else {
		return def
	}

}

func (c *Config) Int(name string, def int) (int, error) {
	if val, ok := c.data[name]; ok {
		return strconv.Atoi(val)
	} else {
		return def, nil
	}
}

func (c *Config) Bool(name string, def bool) (bool, error) {
	if val, ok := c.data[name]; ok {
		val = strings.ToLower(val)

		if val == "true" {
			return true, nil
		} else if val == "false" {
			return false, nil
		} else {
			return false, errors.New("not valid bool value")
		}
	} else {
		return def, nil
	}
}

const spaceChars = " \t\n\v"

func ParseFile(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

func Parse(reader io.Reader) (*Config, error) {
	data := make(map[string]string)
	in := bufio.NewReader(reader)
	ln := 0

	for {
		ln++
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		line = strings.Trim(line, spaceChars)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			return nil, &Error{Line: ln,
				Message: "key=value line format expected"}
		}

		key := strings.Trim(parts[0], spaceChars)
		val := strings.Trim(parts[1], spaceChars)

		data[key] = val
	}

	return &Config{data: data}, nil
}
