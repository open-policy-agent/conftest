package xml

import (
	"bytes"
	"testing"
)

func TestXMLParser(t *testing.T) {
	parser := &Parser{}
	sample := `<note>
	<to>foo</to>
	<from>bar</from>
	<heading>Reminder</heading>
	<body>baz</body>
	</note>`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Fatalf("there should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
	item := inputMap["note"]
	if len(item.(map[string]any)) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
