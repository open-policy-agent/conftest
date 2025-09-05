package toml

import (
	"bytes"
	"testing"
)

func TestTomlParser(t *testing.T) {
	parser := &Parser{}
	sample := `defaultEntryPoints = ["http", "https"]

[entryPoints]
	[entryPoints.http]
	address = ":80"
	compress = true`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Fatalf("there should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
	item := inputMap["entryPoints"]
	if len(item.(map[string]any)) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
