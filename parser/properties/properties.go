package properties

import (
	"encoding/json"
	"fmt"

	prop "github.com/magiconair/properties"
)

// Parser is a properties parser.
type Parser struct{}

func (pp *Parser) Unmarshal(p []byte, v interface{}) error {
	rawProps, err := prop.LoadString(string(p))
	if err != nil {
		return fmt.Errorf("Could not parse properties file: %w", err)
	}

	result := rawProps.Map()

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal properties to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal properties json: %w", err)
	}

	return nil
}
