package ini

import (
	"bytes"
	"fmt"

	"github.com/subosito/gotenv"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
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

	return convert.Remarshal("dotenv", cfg, v)
}
