package properties

import (
	"fmt"

	"encoding/json"
	prop "github.com/magiconair/properties"
)

// Parser is a properties parser.
type Parser struct{}

func (pp *Parser) Unmarshal(p []byte, v interface{}) error {
	raw_props := prop.MustLoadString(string(p))

	result := raw_props.Map()

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal properties to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal properties json: %w", err)
	}

	return nil
}
