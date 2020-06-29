package jsonnet

import (
	"testing"
)

func TestJsonnetParser(t *testing.T) {
	parser := &Parser{}

	// Example from Jsonnet(https://jsonnet.org/)
	sample := `// Edit me!
{
  person1: {
    name: "Alice",
    welcome: "Hello " + self.name + "!",
  },
  person2: self.person1 { name: "Bob" },
}`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	item := input.(map[string]interface{})

	if len(item) <= 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
