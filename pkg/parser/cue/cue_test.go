package cue

import (
	"testing"
	"fmt"
)

func TestCueParser(t *testing.T) {
	parser := &Parser{
		FileName: "sample.cue",
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
	item := inputMap["deployment"]
	fmt.Println(item.(map[string]interface{}))
	if len(item.(map[string]interface{})) <= 0 {
		t.Error("There should be at least one item defined in the parsed file, but none found")
	}
}
