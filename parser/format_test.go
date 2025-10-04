package parser

import (
	"testing"
)

func TestFormat(t *testing.T) {
	t.Run("flattened single config", func(t *testing.T) {
		configurations := map[string][]any{
			"ignored-file-name": {
				map[string]string{"Property": "value"},
			},
		}

		actual, err := Format(configurations)
		if err != nil {
			t.Fatalf("parsing configs: %s", err)
		}

		expected := `{
  "Property": "value"
}`
		if actual != expected {
			t.Errorf("Unexpected formatting. expected '%v' actual '%v'", expected, actual)
		}
	})

	t.Run("multiple configs", func(t *testing.T) {
		configurations := map[string][]any{
			"file1.json": {
				map[string]string{"Sut": "test"},
			},
			"file2.json": {
				map[string]string{"Foo": "bar"},
				map[string]string{"Baz": "cool"},
			},
		}

		actual, err := Format(configurations)
		if err != nil {
			t.Fatalf("format configs: %s", err)
		}

		expected := `{
  "file1.json": {
    "Sut": "test"
  },
  "file2.json": [
    {
      "Foo": "bar"
    },
    {
      "Baz": "cool"
    }
  ]
}`
		if actual != expected {
			t.Errorf("Unexpected formatting. expected '%v' actual '%v'", expected, actual)
		}
	})

}

func TestFormatCombined(t *testing.T) {
	configurations := map[string][]any{
		"file1.json": {
			map[string]string{"Sut": "test"},
		},
		"file2.json": {
			map[string]string{"Foo": "bar"},
		},
	}

	actual, err := FormatCombined(configurations)
	if err != nil {
		t.Fatalf("format configs: %s", err)
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
]`

	if actual != expected {
		t.Errorf("Unexpected combined formatting. expected '%v' actual '%v'", expected, actual)
	}
}
