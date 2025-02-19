package ini

import (
	"testing"
)

func TestIniParser(t *testing.T) {
	parser := &Parser{}
	sample := `[Local Variables]
	Name=name
	Title=title
	Visibility=show/hide
	Delay=10


	[Navigation Controls]
	OnNext=node path
	Help=help file

	# Test comment`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]any)
	item := inputMap["Local Variables"]
	if len(item.(map[string]any)) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}

func TestConvertTypes(t *testing.T) {
	testTable := []struct {
		name           string
		input          map[string]string
		expectedOutput any
	}{
		{"Test number literal", map[string]string{"test": "3.0"}, 3.0},
		{"Test string literal", map[string]string{"test": "conftest"}, "conftest"},
		{"Test boolean literal", map[string]string{"test": "true"}, true},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			val := convertKeyTypes(testUnit.input)
			for _, v := range val {
				if v != testUnit.expectedOutput {
					t.Fatalf("convert type got wrong value %v want %v", v, testUnit.expectedOutput)
				}
			}
		})
	}
}
