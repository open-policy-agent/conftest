package ini

import (
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

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	if len(inputMap) == 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
