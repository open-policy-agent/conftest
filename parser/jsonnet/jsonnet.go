package jsonnet

import (
	"encoding/json"
	"fmt"
	"io"
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

// Parse parses Jsonnet files.
func (p *Parser) Parse(r io.Reader) ([]any, error) {
	vm := jsonnet.MakeVM()
	vm.ErrorFormatter.SetMaxStackTraceSize(20)

	// If path is set, configure import path to the file's directory
	if p.path != "" {
		dir := filepath.Dir(p.path)
		vm.Importer(&jsonnet.FileImporter{
			JPaths: []string{dir},
		})
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	snippetStream, err := vm.EvaluateAnonymousSnippet("", string(data))
	if err != nil {
		return nil, fmt.Errorf("evaluate anonymous snippet: %w", err)
	}

	var v any
	if err := json.Unmarshal([]byte(snippetStream), &v); err != nil {
		return nil, fmt.Errorf("unmarshal json failed: %w", err)
	}

	return []any{v}, nil
}
