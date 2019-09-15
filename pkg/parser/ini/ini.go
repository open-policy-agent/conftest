package ini

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/go-ini/ini"
)

type Parser struct{}

func (i *Parser) Unmarshal(p []byte, v interface{}) error {
	result := map[string]map[string]string{}
	cfg, err := ini.Load(p)
	if err != nil {
		return fmt.Errorf("Fail to read ini file: %v", err)
	}

	sections := cfg.Sections()
	for _, s := range sections {
		sectionName := s.Name()
		if sectionName == "DEFAULT" {
			continue
		}
		result[sectionName] = map[string]string{}
		result[sectionName] = s.KeysHash()
	}
	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Error trying to parse ini to json: %s", err)
	}
	err = yaml.Unmarshal(j, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from ini-json: %s", err)
	}
	return nil
}
