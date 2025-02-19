package parser

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	configurations := make(map[string]any)
	config := struct {
		Property string
	}{
		Property: "value",
	}

	const expectedFileName = "file.json"
	configurations[expectedFileName] = config

	actual, err := Format(configurations)
	if err != nil {
		t.Fatalf("parsing configs: %s", err)
	}

	expected := `
{
	"Property": "value"
}
`

	if !strings.Contains(actual, expected) {
		t.Errorf("Unexpected formatting. expected %v actual %v", expected, actual)
	}

	if !strings.Contains(actual, expectedFileName) {
		t.Errorf("Unexpected formatting. expected %v actual %v", expected, actual)
	}
}

func TestFormatCombined(t *testing.T) {
	configurations := make(map[string]any)
	config := struct {
		Sut string
	}{
		Sut: "test",
	}

	config2 := struct {
		Foo string
	}{
		Foo: "bar",
	}

	configurations["file1.json"] = config
	configurations["file2.json"] = config2

	actual, err := FormatCombined(configurations)
	if err != nil {
		t.Fatalf("parsing configs: %s", err)
	}

	expected := `[
	{
		"path": "file1.json",
		"contents": {
			"Sut": "test"
		}
	},
	{
		"path": "file2.json",
		"contents": {
			"Foo": "bar"
		}
	}
]
`

	if !strings.Contains(actual, expected) {
		t.Errorf("Unexpected combined formatting. expected %v actual %v", expected, actual)
	}
}
