package cue

import (
	"encoding/json"
	"fmt"
	"io"

	"cuelang.org/go/cue/cuecontext"
	cformat "cuelang.org/go/cue/format"
)

// Parser is a CUE parser.
type Parser struct{}

// Parse parses CUE files.
func (*Parser) Parse(r io.Reader) ([]any, error) {
	p, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	out, err := cformat.Source(p)
	if err != nil {
		return nil, fmt.Errorf("format cue: %w", err)
	}

	cueContext := cuecontext.New()
	cueBytes := cueContext.CompileBytes(out)

	cueJSON, err := cueBytes.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	var v any
	if err := json.Unmarshal(cueJSON, &v); err != nil {
		return nil, fmt.Errorf("unmarshal cue json: %w", err)
	}

	return []any{v}, nil
}
