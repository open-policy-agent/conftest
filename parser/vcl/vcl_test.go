package vcl

import (
	"bytes"
	"testing"
)

func TestVCLParser(t *testing.T) {
	parser := &Parser{}
	sample := `acl purge {
	"127.0.0.1";
	"localhost";
}`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	item := input[0].(map[string]any)

	if len(item) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
