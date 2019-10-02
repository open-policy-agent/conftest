package docker_test

import (
	"strings"
	"testing"

	"github.com/instrumenta/conftest/pkg/parser/docker"
)

func TestParser_Unmarshal(t *testing.T) {
	parser := new(docker.Parser)

	sample := `FROM foo
COPY . /
RUN echo hello`

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
