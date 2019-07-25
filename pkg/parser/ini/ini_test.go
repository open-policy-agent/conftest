package ini

import (
	"testing"
)

func TestIniParser(t *testing.T) {
	parser := &Parser{
		FileName: "sample.ini",
	}

	var input interface{}
	err := parser.Unmarshal(nil, &input)
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	item := inputMap["Local Varaibles"]
	if len(item.(map[string]interface{})) <= 0 {
		t.Error("There should be at least one item defined in the parsed file, but none found")
	}
}
