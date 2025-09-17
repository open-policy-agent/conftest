package ini

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/subosito/gotenv"
)

// Parser is an dotenv parser.
type Parser struct{}

// Parse parses dotenv files.
func (i *Parser) Parse(r io.Reader) ([]any, error) {
	cfg, err := gotenv.StrictParse(r)
	if err != nil {
		return nil, fmt.Errorf("read .env file: %w", err)
	}

	j, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal dotenv to json: %w", err)
	}

	var v any
	if err := json.Unmarshal(j, &v); err != nil {
		return nil, fmt.Errorf("unmarshal dotenv json: %w", err)
	}

	return []any{v}, nil
}
