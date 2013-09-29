package server

import (
	"testing"
)

// This command testcases should be parsed successfully.
var passTests = []string{
	"ADD plist path",
	"ADDPLAYLIST Name",
	"LS /",
	"PING",
	"PLAYLIST foo",
	"PLAYLISTS",
	"QUIT",
	"QUIT ",
}

// This command testcases should not be parsed successfully.
var failTests = []string{
	"",
	"ADD",
	"ADD foo",
	"ADD foo bar baz",
	"ADDPLAYLIST",
	"ADDPLAYLIST foo bar",
	"LS",
	"LS foo bar",
	"PING foo",
	"PLAYLIST",
	"PLAYLIST foo bar",
	"PLAYLISTS 1",
	"UNEXISTING_COMMAND",
}

func TestCommand(t *testing.T) {
	for _, cmd := range passTests {
		_, err := parseCommand(cmd)

		if err != nil {
			t.Fatalf("Parsing command '%s' failed. %s",
				cmd, err)
		}
	}

	for _, cmd := range failTests {
		_, err := parseCommand(cmd)
		if err == nil {
			t.Fatalf("'%s' command parsed successfully but should not.",
				cmd)
		}
	}
}
