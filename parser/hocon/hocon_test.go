package hocon

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHoconUnmarshal(t *testing.T) {
	parser := &Parser{}
	sample := `play {
	server {
		dir = ${?user.dir}
	
		# HTTP configuration
		http {
			port = 9001
			port = ${?PLAY_HTTP_PORT}
			port = ${?http.port}
		}
	}
}`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]any)
	item := inputMap["play"]
	if len(item.(map[string]any)) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}

// HOCON allows a scalar to be bound directly at the root of a document, for
// instance `env = prod`. Such keys used to recurse into the value
// unconditionally, which asked the underlying object of a scalar for its keys
// and panicked instead of producing the value.
func TestHoconUnmarshalTopLevelValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:     "top-level string",
			input:    "env = prod\n",
			expected: map[string]any{"env": "prod"},
		},
		{
			name:     "top-level number",
			input:    "port = 8080\n",
			expected: map[string]any{"port": 8080.0},
		},
		{
			name:     "top-level boolean",
			input:    "debug = true\n",
			expected: map[string]any{"debug": true},
		},
		{
			name: "scalars alongside objects",
			input: `env = prod
service {
	port = 9001
}`,
			expected: map[string]any{
				"env": "prod",
				"service": map[string]any{
					"port": 9001.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{}

			var input any
			if err := parser.Unmarshal([]byte(tt.input), &input); err != nil {
				t.Fatalf("parser should not have thrown an error: %v", err)
			}

			if input == nil {
				t.Fatal("there should be information parsed but its nil")
			}

			if diff := cmp.Diff(tt.expected, input); diff != "" {
				t.Errorf("Unmarshal() produced an unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
