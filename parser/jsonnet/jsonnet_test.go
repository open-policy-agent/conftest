package jsonnet

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestJsonnetParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "basic jsonnet with self reference",
			input: `{
				person1: {
					name: "Alice",
					welcome: "Hello " + self.name + "!",
				},
				person2: self.person1 { name: "Bob" },
			}`,
			want: map[string]any{
				"person1": map[string]any{
					"name":    "Alice",
					"welcome": "Hello Alice!",
				},
				"person2": map[string]any{
					"name":    "Bob",
					"welcome": "Hello Bob!",
				},
			},
			wantErr: false,
		},
		{
			name: "arithmetic operations",
			input: `{
				a: 1 + 2,
				b: 6 * 3,
				c: 10 - 5,
				d: 15 / 3,
			}`,
			want: map[string]any{
				"a": float64(3),
				"b": float64(18),
				"c": float64(5),
				"d": float64(5),
			},
			wantErr: false,
		},
		{
			name:    "invalid jsonnet",
			input:   `{ invalid syntax `,
			want:    nil,
			wantErr: true,
			errMsg:  "evaluate anonymous snippet:",
		},
		{
			name: "array and nested objects",
			input: `{
				numbers: [1, 2, 3],
				nested: {
					a: { b: { c: "deep" } },
				},
			}`,
			want: map[string]any{
				"numbers": []any{float64(1), float64(2), float64(3)},
				"nested": map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": "deep",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "stack overflow prevention",
			input: `
				local recurse(x) =
					if x == 0 then
						0
					else
						recurse(x-1) + 1;
				{ result: recurse(1000) }
			`,
			want:    nil,
			wantErr: true,
			errMsg:  "max stack frames exceeded",
		},
	}

	parser := &Parser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			err := parser.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonnetImports(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a library file that will be imported
	libPath := filepath.Join(tmpDir, "lib.libsonnet")
	libContent := `{
  getName(person):: "Hello " + person + "!"
}`
	if err := os.WriteFile(libPath, []byte(libContent), os.FileMode(0600)); err != nil {
		t.Fatalf("failed to write lib file: %v", err)
	}

	// Create main Jsonnet file that imports the library
	mainPath := filepath.Join(tmpDir, "main.jsonnet")
	mainContent := `local lib = import "lib.libsonnet";
{
  greeting: lib.getName("Alice")
}`
	if err := os.WriteFile(mainPath, []byte(mainContent), os.FileMode(0600)); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}

	// Test cases
	tests := []struct {
		name     string
		path     string
		content  []byte
		wantErr  bool
		validate func(t *testing.T, result any)
	}{
		{
			name:    "successful import",
			path:    mainPath,
			content: []byte(mainContent),
			validate: func(t *testing.T, result any) {
				t.Helper()
				m, ok := result.(map[string]any)
				if !ok {
					t.Fatal("result is not a map")
				}
				greeting, ok := m["greeting"].(string)
				if !ok {
					t.Fatal("greeting is not a string")
				}
				if want := "Hello Alice!"; greeting != want {
					t.Errorf("got greeting %q, want %q", greeting, want)
				}
			},
		},
		{
			name:    "import without path set",
			content: []byte(mainContent),
			wantErr: true, // Should fail as import path is not set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{}
			if tt.path != "" {
				parser.SetPath(tt.path)
			}

			var result any
			err := parser.Unmarshal(tt.content, &result)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect success and have a validation function, run it
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
