package properties

import (
	"fmt"

	prop "github.com/magiconair/properties"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Parser is a properties parser.
type Parser struct{}

func (pp *Parser) Unmarshal(p []byte, v any) error {
	rawProps, err := prop.LoadString(string(p))
	if err != nil {
		return fmt.Errorf("parse properties file: %w", err)
	}

	result := rawProps.Map()

	return convert.Remarshal("properties", result, v)
}
