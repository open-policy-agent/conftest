package docker

import (
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	parser := Parser{}

	sample := `FROM foo
COPY . /
RUN echo hello`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	dockerFile := input.([]any)[0]
	commands := dockerFile.([]any)[0]

	expected := "from"
	actual := commands.(map[string]any)["Cmd"]

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

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	dockerFile := input.([]any)[0]
	commands := dockerFile.([]any)

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
