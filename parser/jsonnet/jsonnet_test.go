package jsonnet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJsonnetParser(t *testing.T) {
	parser := &Parser{}

	// Example from Jsonnet(https://jsonnet.org/)
	sample := `// Edit me!
{
  person1: {
    name: "Alice",
    welcome: "Hello " + self.name + "!",
  },
  person2: self.person1 { name: "Bob" },
}`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	item := input.(map[string]interface{})

	if len(item) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
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
		validate func(t *testing.T, result interface{})
	}{
		{
			name:    "successful import",
			path:    mainPath,
			content: []byte(mainContent),
			validate: func(t *testing.T, result interface{}) {
				t.Helper()
				m, ok := result.(map[string]interface{})
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

			var result interface{}
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
