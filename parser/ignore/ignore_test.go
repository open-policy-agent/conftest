package ignore

import (
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	parser := Parser{}

	sample := `!bar

# Test`

	var listOfEntryLists [][]any
	if err := parser.Unmarshal([]byte(sample), &listOfEntryLists); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if listOfEntryLists == nil {
		t.Error("there should be information parsed but it's nil")
	}

	input := listOfEntryLists[0]
	if input == nil {
		t.Error("there should be a list of Entries but it's nil")
	}

	expectedEntryCount := 3

	if len(input) != expectedEntryCount {
		t.Errorf("there should be exactly %v entries in the ignore array but there were %d", expectedEntryCount, len(input))
	}

	firstIgnoreEntry := input[0]

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
