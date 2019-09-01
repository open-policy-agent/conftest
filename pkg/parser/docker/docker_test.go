package docker

import (
	"strings"
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	fileName := "Dockerfile"
	parser := &Parser{
		FileName: fileName,
	}

	sample := `FROM golang:1.12-alpine as builder
	COPY . /
	RUN go build cmd/main.go
	`

	var input interface{}
	err := parser.Unmarshal([]byte(sample), &input)
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	arrayOfCommands := input.([]interface{})
	inputMap := arrayOfCommands[0].(map[string]interface{})
	if strings.Compare(inputMap["Cmd"].(string), "from") != 0 {
		t.Error("The first command should be from")
	}
}
