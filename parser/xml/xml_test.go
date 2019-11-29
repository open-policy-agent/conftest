package xml

import (
	"testing"
)

func TestIniParser(t *testing.T) {
	parser := &Parser{}
	sample := `<note>
	<to>foo</to>
	<from>bar</from>
	<heading>Reminder</heading>
	<body>baz</body>
	</note>`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	item := inputMap["note"]
	if len(item.(map[string]interface{})) <= 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
