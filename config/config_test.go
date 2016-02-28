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
