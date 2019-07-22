package toml

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Parser struct {
	FileName string
}

func (tp *Parser) Unmarshal(p []byte, v interface{}) error {
	err := toml.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("Unable to parse TOML from %s: %s", tp.FileName, err)
	}

	return nil
}
