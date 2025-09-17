package ini

import (
	"bytes"
	"testing"
)

func TestDotenvParser(t *testing.T) {
	parser := &Parser{}
	sample := `MYSQL_HOST_PORT=3307
	MYSQL_IT_HOST_PORT=3308
	MYSQL_ROOT_PASSWORD=root
	MYSQL_DATABASE=root
	MYSQL_USER=root
	MYSQL_PASSWORD=root`

	input, err := parser.Parse(bytes.NewBufferString(sample))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Error("there should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
	if len(inputMap) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
