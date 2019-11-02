package toml

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Parser struct{}

func (tp *Parser) Unmarshal(p []byte, v interface{}) error {
	if err := toml.Unmarshal(p, v); err != nil {
		return fmt.Errorf("unmarshal toml: %w", err)
	}

	return nil
}
