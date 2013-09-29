package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// XXX: Known bug:
//      Standard text/scanner split 111aaa sequense into two tokens (111 & aaa)
//      but it is better to parse it as one. It is not a serious bug, just a
//      little gotcha. I think it can be ignored.

// Available command names.
const (
	cmdAdd         = "ADD"
	cmdAddPlaylist = "ADDPLAYLIST"
	cmdLs          = "LS"
	cmdPing        = "PING"
	cmdPlaylist    = "PLAYLIST"
	cmdPlaylists   = "PLAYLISTS"
	cmdQuit        = "QUIT"
)

// Command argument type type.
type argumentType int

// Supported command argument types.
const (
	argumentTypeString argumentType = iota
	argumentTypeUInt
)

// Command is a object for representation client's action. Client's
// command line is parsed into this struct.
type command struct {
	name string
	args []interface{}
}

// Desctiption of all available commands and their format.
// So we can validate client's requests using this information.
var commandArguments = map[string][]argumentType{
	// Args: playlist name, dir path.
	cmdAdd: []argumentType{argumentTypeString, argumentTypeString},
	// Args: playlist name.
	cmdAddPlaylist: []argumentType{argumentTypeString},
	// Args: dir path.
	cmdLs: []argumentType{argumentTypeString},
	// Args:
	cmdPing: []argumentType{},
	// Args: playlist name.
	cmdPlaylist: []argumentType{argumentTypeString},
	// Args:
	cmdPlaylists: []argumentType{},
	// Args:
	cmdQuit: []argumentType{},
}

// parseCommand converts raw client command-string to the command object.
// Returns an error if input format is not valid.
func parseCommand(str string) (cmd *command, err error) {
	var snr scanner.Scanner

	snr.Init(strings.NewReader(str))
	snr.Whitespace = 1<<'\t' | 1<<' '
	snr.Mode = scanner.ScanInts | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanIdents

	var tokens []string

	next := snr.Scan()
	for next != scanner.EOF {
		tokens = append(tokens, strings.Trim(snr.TokenText(), "\""))
		next = snr.Scan()
	}

	if len(tokens) == 0 {
		return nil, errors.New("Illegal command format.")
	}

	cmd = new(command)
	cmd.name = tokens[0]

	expectedArgTypes, ok := commandArguments[cmd.name]
	if !ok {
		return nil, fmt.Errorf("Unsupported command %s.", cmd.name)
	}

	args := tokens[1:]
	if len(args) != len(expectedArgTypes) {
		return nil, fmt.Errorf("Command %s expects %d parameters but %d given.",
			cmd.name, len(expectedArgTypes), len(args))
	}

	// Convert list of parsed tokens to the command object arguments.
	for i := 0; i < len(args); i++ {
		arg := args[i]
		expArgType := expectedArgTypes[i]

		var a interface{}

		switch expArgType {
		case argumentTypeString:
			// Any type can be used as a string, so just
			// use it without any validation or convertion.
			a = arg
		case argumentTypeUInt:
			n, err := strconv.Atoi(arg)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("%s cannot be converted to unsigned int", arg)
			}
			a = n
		default:
			panic("Unexpected argument type found.")
		}

		cmd.args = append(cmd.args, a)
	}

	return cmd, nil
}
