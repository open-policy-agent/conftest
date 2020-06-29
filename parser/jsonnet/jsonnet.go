package jsonnet

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-jsonnet"
)

// Parser is a Jsonnet parse
type Parser struct{}


// Unmarshal unmarshals Jsonnet files
func (p *Parser) Unmarshal(data []byte, v interface{}) error {
	vm := jsonnet.MakeVM()
	snippetStream, err := vm.EvaluateSnippet("", string(data))
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(snippetStream), v); err != nil {
		return fmt.Errorf("unmarshal json failed: %w", err)
	}

	return nil
}
