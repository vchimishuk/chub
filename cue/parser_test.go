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

package cue

import (
	"testing"
)

type expected struct {
	Cmd    string
	Params []string
}

type test struct {
	Input  string
	Etalon expected
}

func TestParseCommand(t *testing.T) {
	var tests = []test{
		{"COMMAND",
			expected{"COMMAND",
				[]string{}}},
		{"COMMAND \t PARAM1   PARAM2\tPARAM3",
			expected{"COMMAND",
				[]string{"PARAM1", "PARAM2", "PARAM3"}}},
		{"COMMAND 'PARAM1' \"PARAM2\" 'PAR\"AM3' 'P AR  AM 4'",
			expected{"COMMAND",
				[]string{"PARAM1", "PARAM2", "PAR\"AM3", "P AR  AM 4"}}},
		{"COMMAND 'P A R A M 1' \"PA RA M2\" PA\\\"RAM\\'3",
			expected{"COMMAND",
				[]string{"P A R A M 1", "PA RA M2", "PA\"RAM'3"}}},
	}

	for _, tt := range tests {
		cmd, params, err := parseCommand(tt.Input)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if cmd != tt.Etalon.Cmd {
			t.Fatalf("Parsed command '%s' but '%s' expected", cmd, tt.Etalon.Cmd)
		}

		if len(params) != len(tt.Etalon.Params) {
			t.Fatalf("Parsed %d params but %d expected", len(params), len(tt.Etalon.Params))
		}

		for i := 0; i < len(params); i++ {
			if params[i] != tt.Etalon.Params[i] {
				t.Fatalf("Parsed '%s' parameter but '%s' expected", params[i], tt.Etalon.Params[i])
			}
		}
	}
}

type timeExpected struct {
	min    int
	sec    int
	frames int
}

func TestParseTime(t *testing.T) {
	var tests = map[string]timeExpected{
		"01:02:03": timeExpected{1, 2, 3},
		"11:22:33": timeExpected{11, 22, 33},
		"14:00:00": timeExpected{14, 0, 0},
	}

	for input, expected := range tests {
		min, sec, frames, err := parseTime(input)
		if err != nil {
			t.Fatalf("Time parsing failed. Input string: '%s'. %", input, err)
		}

		if min != expected.min {
			t.Fatalf("Expected %d minutes, but %d recieved.", expected.min, min)
		}
		if sec != expected.sec {
			t.Fatalf("Expected %d seconds, but %d recieved.", expected.sec, sec)
		}
		if frames != expected.frames {
			t.Fatalf("Expected %d frames, but %d recieved.", expected.frames, frames)
		}
	}
}
