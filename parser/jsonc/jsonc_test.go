package jsonc

import (
	"testing"
)

func TestJSONParser(t *testing.T) {
	parser := &Parser{}
	sample := `{
  "name": "conftest-example", // Ignore comments
  "version": "1.0.0",
  "description": "An example of testing Typescript code with Open Policy Agent",
  "main": "pod.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  /* Like this */
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
