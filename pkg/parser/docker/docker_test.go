package docker

import (
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {

	parser := Parser{}

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

	dockerFile := input.([]interface{})[0]
	commands := dockerFile.([]interface{})[0]

	expected := "from"
	actual := commands.(map[string]interface{})["Cmd"]

	if actual != expected {
		t.Errorf("First Docker command should be '%v', was '%v'", expected, actual)
	}
}
