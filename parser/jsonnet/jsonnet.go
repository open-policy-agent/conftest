package jsonnet

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-jsonnet"
)

// Parser is a Jsonnet parser.
type Parser struct{}

// Unmarshal unmarshals Jsonnet files.
func (p *Parser) Unmarshal(data []byte, v interface{}) error {
	vm := jsonnet.MakeVM()
	vm.ErrorFormatter.SetMaxStackTraceSize(20)
	snippetStream, err := vm.EvaluateAnonymousSnippet("", string(data))
	if err != nil {
		return fmt.Errorf("evaluate anonymous snippet: %w", err)
	}

	if err := json.Unmarshal([]byte(snippetStream), v); err != nil {
		return fmt.Errorf("unmarshal json failed: %w", err)
	}

	return nil
}
