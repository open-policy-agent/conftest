package commands

import (
	"strings"
	"testing"
)

func TestParse_ByDefault_AddsIndentationAndNewline(t *testing.T) {
	configurations := make(map[string]interface{})

	config := struct {
		Property string
	}{
		Property: "value",
	}

	const expectedFileName = "file.json"
	configurations[expectedFileName] = config

	actual, err := parseConfigurations(configurations)
	if err != nil {
		t.Fatalf("parsing configs: %s", err)
	}

	expected := `
{
	"Property": "value"
}
`

	if !strings.Contains(actual, expected) {
		t.Errorf("unexpected parsed config. expected %v actual %v", expected, actual)
	}

	if !strings.Contains(actual, expectedFileName) {
		t.Errorf("unexpected parsed filename. expected %v actual %v", expected, actual)
	}
}
