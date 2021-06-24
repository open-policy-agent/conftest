package properties

import (
	"testing"
)

func TestPropertiesParser(t *testing.T) {
	parser := &Parser{}
	sample := `# This is a simle properties file
    SAMPLE_KEY=https://example.com/
! some comment
my-property=some-value`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Fatalf("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	myProp := inputMap["my-property"].(string)
	if myProp != "some-value" {
		t.Fatalf("Failed to parse property: %s", myProp)
	}

	spaceProp := inputMap["SAMPLE_KEY"].(string)
	if spaceProp != "https://example.com/" {
		t.Fatalf("Failed to strip whitespace from key: %s", myProp)
	}
}
