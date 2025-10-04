package docker

import (
	"bytes"
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	parser := Parser{}

	sample := `FROM foo
COPY . /
RUN echo hello`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Error("there should be information parsed but its nil")
	}

	commands := input[0].([]any)
	if len(commands) != 3 {
		t.Error("there should be three commands parsed")
	}

	expected := "from"
	actual := commands[0].(map[string]any)["Cmd"]

	if actual != expected {
		t.Errorf("first Docker command should be '%v', was '%v'", expected, actual)
	}
}

func TestParser_Unmarshal_Multistage(t *testing.T) {
	parser := Parser{}

	sample := `FROM golang:1.13-alpine as base
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
	
FROM base as builder
RUN go build -o conftest`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Error("there should be information parsed but its nil")
	}

	commands := input[0].([]any)

	cmd := commands[1]
	stage := cmd.(map[string]any)["Stage"].(float64)
	if stage != 0 {
		t.Errorf("expected command to be in stage 0, not stage: %v", stage)
	}

	cmd = commands[6]
	stage = cmd.(map[string]any)["Stage"].(float64)
	if stage != 1 {
		t.Errorf("expected command to be in stage 1, not stage: %v", stage)
	}
}
