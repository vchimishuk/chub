package cmd

import "testing"

func TestString(t *testing.T) {
	s := newScanner(" те\"ст \"'юнікоду'\" ")

	if !s.HasNext() {
		t.Fatal()
	}
	str, err := s.NextString()
	if err != nil || str != "те\"ст" {
		t.Fatal()
	}
	if !s.HasNext() {
		t.Fatal()
	}
	str, err = s.NextString()
	if err != nil || str != "'юнікоду'" {
		t.Fatal()
	}
	if s.HasNext() {
		t.Fatal()
	}
}

func TestPath(t *testing.T) {
	s := newScanner("/foo\\ bar/baz")

	if !s.HasNext() {
		t.Fatal()
	}
	str, err := s.NextString()
	if err != nil || str != "/foo bar/baz" {
		t.Fatal()
	}
	if s.HasNext() {
		t.Fatal()
	}
}

func TestInt(t *testing.T) {
	s := newScanner("+123 -321 0")

	if !s.HasNext() {
		t.Fatal()
	}
	n, err := s.NextInt()
	if err != nil || n != 123 {
		t.Fatal()
	}
	if !s.HasNext() {
		t.Fatal()
	}
	n, err = s.NextInt()
	if err != nil || n != -321 {
		t.Fatal()
	}
	if !s.HasNext() {
		t.Fatal()
	}
	n, err = s.NextInt()
	if err != nil || n != 0 {
		t.Fatal()
	}
	if s.HasNext() {
		t.Fatal()
	}
}

func TestCommand(t *testing.T) {
	s := newScanner(`PLAYLIST_APPEND "Test" "/home/viacheslav/documents/music"`)

	if !s.HasNext() {
		t.Fatal()
	}
	str, err := s.NextString()
	if err != nil || str != "PLAYLIST_APPEND" {
		t.Fatal()
	}
	if !s.HasNext() {
		t.Fatal()
	}
	str, err = s.NextString()
	if err != nil || str != "Test" {
		t.Fatal()
	}
	if !s.HasNext() {
		t.Fatal()
	}
	str, err = s.NextString()
	if err != nil || str != "/home/viacheslav/documents/music" {
		t.Fatal()
	}
	if s.HasNext() {
		t.Fatal()
	}
}
