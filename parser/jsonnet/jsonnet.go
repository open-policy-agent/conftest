package jsonnet

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/google/go-jsonnet"
)

// Parser is a Jsonnet parser.
type Parser struct {
	path string
}

// SetPath sets the original file path for relative imports
func (p *Parser) SetPath(path string) {
	p.path = path
}

// Unmarshal unmarshals Jsonnet files.
func (p *Parser) Unmarshal(data []byte, v any) error {
	vm := jsonnet.MakeVM()
	vm.ErrorFormatter.SetMaxStackTraceSize(20)

	// If path is set, configure import path to the file's directory
	if p.path != "" {
		dir := filepath.Dir(p.path)
		vm.Importer(&jsonnet.FileImporter{
			JPaths: []string{dir},
		})
	}

	snippetStream, err := vm.EvaluateAnonymousSnippet("", string(data))
	if err != nil {
		return fmt.Errorf("evaluate anonymous snippet: %w", err)
	}

	if err := json.Unmarshal([]byte(snippetStream), v); err != nil {
		return fmt.Errorf("unmarshal json failed: %w", err)
	}

	return nil
}
