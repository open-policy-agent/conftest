package cue

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue/cuecontext"
	cformat "cuelang.org/go/cue/format"
)

// Parser is a CUE parser.
type Parser struct{}

// Unmarshal unmarshals CUE files.
func (*Parser) Unmarshal(p []byte, v interface{}) error {
	out, err := cformat.Source(p)
	if err != nil {
		return fmt.Errorf("format cue: %w", err)
	}

	cueContext := cuecontext.New()
	cueBytes := cueContext.CompileBytes(out)

	cueJSON, err := cueBytes.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err := json.Unmarshal(cueJSON, v); err != nil {
		return fmt.Errorf("unmarshal cue json: %w", err)
	}

	return nil
}
