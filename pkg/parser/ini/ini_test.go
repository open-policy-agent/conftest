package ini

import (
	"testing"
)

func TestIniParser(t *testing.T) {
	parser := &Parser{}
	sample := `[Local Varaibles] 
	Name=name 
	Title=title 
	Visisbility=show/hide
	
	[Navigation Controls] 
	OnNext=node path 
	Help=help file
	
	# Test comment`

	var input interface{}
	err := parser.Unmarshal([]byte(sample), &input)
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
