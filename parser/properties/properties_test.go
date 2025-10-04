package properties

import (
	"bytes"
	"testing"
)

func TestPropertiesParser(t *testing.T) {
	parser := &Parser{}
	sample := `# This is a simle properties file
    SAMPLE_KEY=https://example.com/
! some comment=not-a-prop
my-property=some-value`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Errorf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Errorf("there should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
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
