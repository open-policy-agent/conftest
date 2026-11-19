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

func TestParser_Unmarshal_TrailingBackslash(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedKind string
		expectedVal  string
	}{
		{"path", `a: b\`, "Path", "a: b"},
		{"path with space before backslash", `a: b \`, "Path", "a: b "},
		{"escaped backslash", `a\\`, "Path", `a\`},
		{"only a backslash", `\`, "Path", ""},
		{"negated path", `!foo\`, "NegatedPath", "foo"},
		{"comment", `# foo\`, "Comment", `foo\`},
		{"escaped trailing space still works", `a\ `, "Path", "a "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got [][]map[string]any
			if err := (&Parser{}).Unmarshal([]byte(tt.input+"\nnext\n"), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if len(got) != 1 || len(got[0]) != 3 {
				t.Fatalf("expected 3 entries, got %v", got)
			}

			e := got[0][0]
			if e["Kind"] != tt.expectedKind || e["Value"] != tt.expectedVal || e["Original"] != tt.input {
				t.Errorf("got %v, want Kind=%q Value=%q Original=%q", e, tt.expectedKind, tt.expectedVal, tt.input)
			}

			if got[0][1]["Value"] != "next" {
				t.Errorf("line after the backslash line was not parsed: %v", got[0][1])
			}
		})
	}
}
