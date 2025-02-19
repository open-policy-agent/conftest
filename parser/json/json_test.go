package json

import (
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

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Fatalf("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]any)
	if len(inputMap) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}

func TestJSONParserWithBOM(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "valid JSON with BOM",
			input: append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"test": "value"}`)...),
			want:  map[string]any{"test": "value"},
		},
		{
			name:  "valid JSON without BOM",
			input: []byte(`{"test": "value"}`),
			want:  map[string]any{"test": "value"},
		},
	}

	parser := &Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			err := parser.Unmarshal(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Fatal("expected parsed content, got nil")
			}
			if m, ok := got.(map[string]any); ok {
				for k, want := range tt.want {
					if got := m[k]; got != want {
						t.Errorf("key %q = %v, want %v", k, got, want)
					}
				}
			}
		})
	}
}
