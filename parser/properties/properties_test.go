package properties

import (
	"testing"
)

func TestPropertiesParser(t *testing.T) {
	parser := &Parser{}
	sample := `# This is a simle properties file
    SAMPLE_KEY=https://example.com/
! some comment=not-a-prop
my-property=some-value`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Errorf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Errorf("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	myProp := inputMap["my-property"].(string)
	if myProp != "some-value" {
		t.Errorf("Failed to parse property: %s", myProp)
	}

	spaceProp := inputMap["SAMPLE_KEY"].(string)
	if spaceProp != "https://example.com/" {
		t.Errorf("Failed to strip whitespace from key: %s", myProp)
	}

	inputLen := len(inputMap)
	if inputLen != 2 {
		t.Errorf("Failed to parse all properties: expected 2 got %d", inputLen)
	}
}
