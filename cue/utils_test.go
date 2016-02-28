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

package cue

import (
	"testing"
)

func TestUtils(t *testing.T) {
	const input string = "1234567890"
	var inputLen = len(input)

	for i := 0; i < inputLen+2; i++ {
		out := stringTruncate(input, i)
		l := len(out)
		if i < inputLen {
			if l != i {
				t.Fatalf("Assertion failed: len(stringTruncate(\"%s\", %d)) == %d", input, i, l)
			}
		} else {
			if l != inputLen {
				t.Fatalf("Assertion failed: len(stringTruncate(\"%s\", %d)) == %d", input, i, inputLen)
			}
		}
	}
}
