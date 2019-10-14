package toml

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Parser struct{}

func (tp *Parser) Unmarshal(p []byte, v interface{}) error {
	err := toml.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("Unable to parse TOML: %s", err)
	}

	return nil
}
