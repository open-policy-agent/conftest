package ignore

import (
	"encoding/json"
	"fmt"
	"io"

	ignore "github.com/shteou/go-ignore"
)

// Parser is a ignore (dockerignore, gitignore) parser.
type Parser struct{}

// Parse parses ignore files.
func (pp *Parser) Parse(r io.Reader) ([]any, error) {
	p, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	ignoreEntries, err := ignore.ParseIgnoreBytes(p)
	if err != nil {
		return nil, fmt.Errorf("parse ignore bytes: %w", err)
	}

	marshalledLines, err := json.Marshal(ignoreEntries)
	if err != nil {
		return nil, fmt.Errorf("marshal ignore: %w", err)
	}

	var v any
	if err := json.Unmarshal(marshalledLines, &v); err != nil {
		return nil, fmt.Errorf("unmarshal ignore: %w", err)
	}

	return []any{v}, nil
}
