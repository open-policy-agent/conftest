package vcl

import (
	"encoding/json"
	"fmt"

	"github.com/KeisukeYamashita/go-vcl/vcl"
)

type Parser struct{}

func (p *Parser) Unmarshal(b []byte, v interface{}) error {
	result := make(map[string]interface{})
	if errs := vcl.Decode(b, &result); len(errs) > 0 {
		return errs[0]
	}

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal vcl to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal vcl json: %w", err)
	}

	return nil
}
