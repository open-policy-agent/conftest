package parser

import (
	"testing"
)

func TestInputWithinputParam(t *testing.T) {
	input := &Input{
		input: "ini",
		fName: "-",
	}

	i := GetInput("-", "ini")
	if i.input != input.input {
		t.Error("input should be ini with given input param")
	}
}

func TestInputWithFileNameParam(t *testing.T) {
	input := &Input{
		input: "",
		fName: "gke.tf",
	}

	i := GetInput("gke.tf", "")
	if i.fName != input.fName {
		t.Error("fileName should passed to object")
	}
	if i.input != "tf" {
		t.Error("input should detected tf with filename suffix")
	}
}
