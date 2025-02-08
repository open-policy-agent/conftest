package ini

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/subosito/gotenv"
)

// Parser is an dotenv parser.
type Parser struct{}

// Unmarshal unmarshals dotenv files.
func (i *Parser) Unmarshal(p []byte, v any) error {
	r := bytes.NewReader(p)
	cfg, err := gotenv.StrictParse(r)
	if err != nil {
		return fmt.Errorf("read .env file: %w", err)
	}

	j, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal dotenv to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal dotenv json: %w", err)
	}

	return nil
}
