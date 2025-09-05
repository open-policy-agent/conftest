package ignore

import (
	"bytes"
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	parser := Parser{}

	sample := `!bar

# Test`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Error("there should be information parsed but it's nil")
	}

	entries := input[0].([]any)
	expectedEntryCount := 3
	if len(entries) != expectedEntryCount {
		t.Errorf("there should be exactly %v entries in the ignore array but there were %d", expectedEntryCount, len(input))
	}

	firstIgnoreEntry := entries[0]

	expectedKind := "NegatedPath"
	actualKind := firstIgnoreEntry.(map[string]any)["Kind"]

	if actualKind != expectedKind {
		t.Errorf("first ignore entry's Kind should be '%v', was '%v'", expectedKind, actualKind)
	}

	expectedValue := "bar"
	actualValue := firstIgnoreEntry.(map[string]any)["Value"]

	if actualValue != expectedValue {
		t.Errorf("first ignore entry's Value should be '%v', was '%v'", expectedValue, actualValue)
	}

	expectedOriginal := "!bar"
	actualOriginal := firstIgnoreEntry.(map[string]any)["Original"]

	if actualOriginal != expectedOriginal {
		t.Errorf("first ignore entry's Kind should be '%v', was '%v'", expectedOriginal, actualOriginal)
	}
}
