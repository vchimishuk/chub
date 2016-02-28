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
