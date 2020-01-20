package commands

import (
	"strings"
	"testing"
)

func TestParse_ByDefault_AddsIndentationAndNewline(t *testing.T) {
	combine := false
	configurations := make(map[string]interface{})

	config := struct {
		Property string
	}{
		Property: "value",
	}

	const expectedFileName = "file.json"
	configurations[expectedFileName] = config

	actual, err := parseConfigurations(configurations, combine)
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

func TestParse_MultiFileCombineFlag(t *testing.T) {
	combine := true
	configurations := make(map[string]interface{})

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

	actual, err := parseConfigurations(configurations, combine)
	if err != nil {
		t.Fatalf("parsing configs: %s", err)
	}

	expected := `{
	"file1.json": {
		"Sut": "test"
	},
	"file2.json": {
		"Foo": "bar"
	}
}
`

	if !strings.Contains(actual, expected) {
		t.Errorf("unexpected parsed config. expected %v actual %v", expected, actual)
	}
}
