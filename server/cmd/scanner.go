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
	"io"
	"strconv"
	"strings"
	"unicode"
)

type scanner struct {
	reader *strings.Reader
}

func newScanner(s string) *scanner {
	return &scanner{reader: strings.NewReader(s)}
}

func (s *scanner) HasNext() bool {
	s.eatSpaces()

	return s.reader.Len() > 0
}

// TODO: Only double quoted strings is supported now,
//       add single quoted strings support too.
func (s *scanner) NextString() (string, error) {
	s.eatSpaces()

	quoted := false
	str := ""
	for {
		r, _, err := s.reader.ReadRune()

		if err != nil {
			if quoted {
				return "", errors.New("unexpeted end of string")
			}
			break
		}
		if len(str) == 0 && r == '"' {
			quoted = true
		} else if r == '\\' {
			r, _, err := s.reader.ReadRune()
			if err != nil {
				return "", errors.New("unknown escape sequence: EOF")
			}
			str += string(r)
		} else if !quoted && unicode.IsSpace(r) {
			break
		} else if quoted && r == '"' {
			break
		} else {
			str += string(r)
		}
	}

	if len(str) == 0 {
		return "", io.EOF
	}

	return str, nil
}

func (s *scanner) NextInt() (int, error) {
	s.eatSpaces()

	first := true
	str := ""
	for {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			break // io.EOF
		}
		if first && (r == '-' || r == '+') {
			str += string(r)
		} else if unicode.IsNumber(r) {
			str += string(r)
		} else if unicode.IsSpace(r) {
			break
		} else {
			return 0, errors.New("invalid integer format")
		}

		first = false
	}

	n, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.New("invalid integer format")
	}

	return n, nil
}

func (s *scanner) eatSpaces() {
	for {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			break // io.EOF
		}
		if !unicode.IsSpace(r) {
			s.reader.UnreadRune()
			break
		}
	}
}
