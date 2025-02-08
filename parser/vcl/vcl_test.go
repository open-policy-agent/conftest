package vcl

import "testing"

func TestVCLParser(t *testing.T) {
	parser := &Parser{}
	sample := `acl purge {
	"127.0.0.1";
	"localhost";
}`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	item := input.(map[string]any)

	if len(item) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
