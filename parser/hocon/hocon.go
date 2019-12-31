package hocon

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-akka/configuration"
	"github.com/go-akka/configuration/hocon"
)

type Parser struct{}

func (i *Parser) Unmarshal(p []byte, v interface{}) error {
	rootCfg := configuration.ParseString(string(p))
	result := make(map[string]interface{})

	for _, key := range rootCfg.Root().GetObject().GetKeys() {
		cfg := rootCfg.GetConfig(key)
		result[key] = getConfig(rootCfg, cfg, key)
	}

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal hocon to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal hocon json: %w", err)
	}

	return nil
}

func getConfig(rootCfg, cfg *configuration.Config, path string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, key := range cfg.Root().GetObject().GetKeys() {
		tmpKey := path + "." + key
		if rootCfg.IsObject(tmpKey) {
			result[key] = getConfig(rootCfg, rootCfg.GetConfig(tmpKey), tmpKey)
		} else {
			value := rootCfg.GetValue(tmpKey)
			result[key] = convertType(value)
		}
	}

	return result
}

func convertType(value *hocon.HoconValue) interface{} {
	str := value.String()
	switch {
	case isNumberLiteral(str):
		num, _ := strconv.ParseFloat(str, 64)
		return num
	case isBooleanLiteral(str):
		b, _ := strconv.ParseBool(str)
		return b
	default:
		return str
	}
}

func isNumberLiteral(f string) bool {
	_, err := strconv.ParseFloat(f, 64)
	return err == nil
}

func isBooleanLiteral(b string) bool {
	_, err := strconv.ParseBool(b)
	return err == nil
}
