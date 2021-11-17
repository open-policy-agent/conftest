package toml

import (
	"testing"
)

func TestTomlParser(t *testing.T) {
	parser := &Parser{}
	sample := `defaultEntryPoints = ["http", "https"]

[entryPoints]
	[entryPoints.http]
	address = ":80"
	compress = true`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Fatalf("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	item := inputMap["entryPoints"]
	if len(item.(map[string]interface{})) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
