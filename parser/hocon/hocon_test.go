package hocon

import "testing"

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
