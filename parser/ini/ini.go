package ini

import (
	"encoding/json"
	"fmt"

	"github.com/go-ini/ini"
)

type Parser struct{}

func (i *Parser) Unmarshal(p []byte, v interface{}) error {
	cfg, err := ini.Load(p)
	if err != nil {
		return fmt.Errorf("read ini file: %w", err)
	}

	result := make(map[string]map[string]string)
	for _, s := range cfg.Sections() {
		sectionName := s.Name()
		if sectionName == "DEFAULT" {
			continue
		}

		result[sectionName] = map[string]string{}
		result[sectionName] = s.KeysHash()
	}

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal ini to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal ini json: %w", err)
	}

	return nil
}
