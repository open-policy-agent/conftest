package docker

import (
	"strings"
	"io/ioutil"
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	fileName := "Dockerfile"
	parser := &Parser{
		FileName: fileName,
	}

	bytes, _ := ioutil.ReadFile(fileName)
	
	var input interface{}
	err := parser.Unmarshal(bytes, &input)
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
