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
