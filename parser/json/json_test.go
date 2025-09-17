package json

import (
	"bytes"
	"reflect"
	"testing"
)

func TestJSONParser(t *testing.T) {
	parser := &Parser{}
	sample := `{
  "name": "conftest-example",
  "version": "1.0.0",
  "description": "An example of testing Typescript code with Open Policy Agent",
  "main": "pod.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "author": "",
  "license": "ISC",
  "dependencies": {
    "js-yaml": "^3.13.1",
    "kubernetes-types": "^1.13.0-beta.1"
  },
  "devDependencies": {
    "ts-node": "^8.1.0",
    "typescript": "^3.4.5"
  }
}`

	input, err := parser.Parse(bytes.NewReader([]byte(sample)))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Fatalf("there should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
	if len(inputMap) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}

func TestJSONParserWithBOM(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []any
		wantErr bool
	}{
		{
			name:  "valid JSON with BOM",
			input: append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"test": "value"}`)...),
			want:  []any{map[string]any{"test": "value"}},
		},
		{
			name:  "valid JSON without BOM",
			input: []byte(`{"test": "value"}`),
			want:  []any{map[string]any{"test": "value"}},
		},
	}

	parser := &Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(bytes.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Fatal("expected parsed content, got nil")
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Expected\n%T : %v\n to equal\n%T : %v\n", got, got, tt.want, tt.want)
			}
		})
	}
}
