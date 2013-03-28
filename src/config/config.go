// Application configuration file parser implementation and some other
// configuration related utility functions.
//
// Configuration file expected to be located at ~/.chub/config in traditional
// UNIX key-value plain text file. E.g.
// foo.bar = "some value"
// foo.baz = "some another value"
// bar = 14
package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strings"
)

// Space characters list.
const space = " \t\n\v"

// Available configuration option types.
const (
	typeString = iota
	typeInt
	typeBool
	typeEnum
)

// Available configuration option names.
const (
	outputName = "output.name"
	vfsRoot    = "vfs.root"
)

// Configuration is a collection of configuration parameters from parsed file.
// Any application code should read this parameters form this object.
var Configurations ConfigurationsMap

// parserErorr represents file parsing errors.
type parserErorr struct {
	msg  string
	line int
	col  int
}

// newParserError creates new parser error object.
func newParserError(msg string, line int, col int) *parserErorr {
	return &parserErorr{msg, line, col}
}

// Error returns string representation of the error object.
func (e *parserErorr) Error() string {
	return fmt.Sprintf("Line: %d, Col: %d. %s", e.line, e.col, e.msg)
}

// ConfigurationsMap interface provides methods for reading configuration
// parameters.
type ConfigurationsMap interface {
	// OutputName returns default output name which should be used for
	// playback.
	OutputName() string
	// VfsRoot returns root folder for VFS, where application is chrooted
	// and can use FS upper this folder.
	VfsRoot() string
}

// entry represents one configuration option record.
type entry struct {
	// Option value type.
	Type int
	// Parse parses option value from string or rises error on failure.
	Parse func(v string) (val interface{}, err error)
	// Value contains parsed value inself.
	Value interface{}
}

// configurationsMap is used for providing ConfigurationsMap.
type configurationsMap struct {
	Entries map[string]entry
}

func (m *configurationsMap) stringVal(name string) string {
	return m.Entries[name].Value.(string)
}

func (m *configurationsMap) intVal(name string) int {
	return m.Entries[name].Value.(int)
}

func (m *configurationsMap) boolVal(name string) bool {
	return m.Entries[name].Value.(bool)
}

func (m *configurationsMap) OutputName() string {
	return m.stringVal(outputName)
}

func (m *configurationsMap) VfsRoot() string {
	return m.stringVal(vfsRoot)
}

func parseOutputName(raw string) (val interface{}, err error) {
	if raw != "alsa" {
		return nil, fmt.Errorf("Unsupported output %s.", raw)
	}

	return raw, nil
}

func parseVfsRoot(raw string) (val interface{}, err error) {
	raw = path.Clean(raw)
	// Open doesn't understand path started with ~ character,
	// so make such path absolute.
	if strings.HasPrefix(raw, "~") {
		raw = path.Join(getHomeDir(), raw[1:])
	}

	dir, err := os.Open(raw)
	if err != nil {
		return nil, err
	}
	fi, err := dir.Stat()
	if !fi.IsDir() {
		return nil, fmt.Errorf("Folder expected.")
	}

	return raw, nil
}

// parser parses given configuration file and returs map filled with its content.
func parse(filename string) (conf ConfigurationsMap, err error) {
	cnf := &configurationsMap{Entries: map[string]entry{
		outputName: {Type: typeString, Parse: parseOutputName, Value: "alsa"},
		vfsRoot:    {Type: typeString, Parse: parseVfsRoot, Value: "/"}}}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	in := bufio.NewReader(file)

	lineNumber := 0
	for {
		lineNumber++
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		line = strings.Trim(line, space)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			err = newParserError("'key=value' line format expected, but simple string found.",
				lineNumber, 0)

			return nil, err
		}

		key := strings.Trim(parts[0], space)
		val := strings.Trim(parts[1], space)

		opt, ok := cnf.Entries[key]
		if !ok {
			e := fmt.Sprintf("Unsupported configuration key %s.", key)
			err = newParserError(e, lineNumber, 0)

			return nil, err
		}

		v, err := opt.Parse(val)
		if err != nil {
			e := fmt.Sprintf("Invalid value %s: %s", val, err)
			err = newParserError(e, lineNumber, 0)

			return nil, err
		}
		opt.Value = v
		cnf.Entries[key] = opt
	}

	return cnf, nil
}

// getHomeDir returns absolute path to the current user's home directory folder.
func getHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Can't retrieve current user information. %s", err))
	}

	return u.HomeDir
}

// Module initialization function.
func init() {
	cf := path.Join(getHomeDir(), ".chub/config")
	conf, err := parse(cf)
	if err != nil {
		panic(fmt.Sprintf("Configuration file test.txt parsing failed. %s\n", err))
	}
	Configurations = conf
}
