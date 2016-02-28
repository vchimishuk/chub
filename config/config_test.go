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

package config

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	input := `
# A comment.
foo = foo
bar = a bar
`

	cfg, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Defined("foo") || !cfg.Defined("bar") {
		t.Fatal()
	}
	if cfg.String("foo", "") != "foo" {
		t.Fatal()
	}
	if cfg.String("bar", "") != "a bar" {
		t.Fatal()
	}
	if def := cfg.String("undefined", "def"); def != "def" {
		t.Fatal()
	}
}

func TestBool(t *testing.T) {
	input := `
foo = true
bar = false
baz = not a bool
`

	cfg, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Defined("foo") || !cfg.Defined("bar") || !cfg.Defined("baz") {
		t.Fatal()
	}
	if b, err := cfg.Bool("foo", false); b != true || err != nil {
		t.Fatal()
	}
	if b, err := cfg.Bool("bar", true); b != false || err != nil {
		t.Fatal()
	}
	if _, err := cfg.Bool("baz", false); err == nil {
		t.Fatal()
	}
}

func TestInt(t *testing.T) {
	input := `
foo = 123
bar = baz
`

	cfg, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Defined("foo") {
		t.Fatal()
	}
	if i, err := cfg.Int("foo", 0); i != 123 || err != nil {
		t.Fatal()
	}
	if _, err := cfg.Int("bar", 0); err == nil {
		t.Fatal()
	}
}
