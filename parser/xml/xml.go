package xml

import (
	"bytes"
	"encoding/json"
	"fmt"

	x "github.com/basgys/goxml2json"
)

type Parser struct{}

func (xml *Parser) Unmarshal(p []byte, v interface{}) error {
	res, err := x.Convert(bytes.NewReader(p))
	if err != nil {
		return fmt.Errorf("unmarshal xml: %w", err)
	}

	if err := json.Unmarshal(res.Bytes(), v); err != nil {
		return fmt.Errorf("convert xml to json: %w", err)
	}

	return nil
}
