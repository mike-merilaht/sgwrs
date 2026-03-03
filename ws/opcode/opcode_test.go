package opcode

import (
	"testing"
)

func TestString(t *testing.T) {
	opcode := OpCodeText
	want := OpCodeStrings[1]
	if opcode.String() != want {
		t.Errorf(`Incorrect string conversion for OpCodeText = %s, expected %s`, opcode.String(), want)
	}
}