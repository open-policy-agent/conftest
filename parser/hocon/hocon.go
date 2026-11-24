package hocon

import (
	"strconv"

	"github.com/go-akka/configuration"
	"github.com/go-akka/configuration/hocon"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Parser is a HOCON parser.
type Parser struct{}

// Unmarshal unmarshals HOCON files.
func (i *Parser) Unmarshal(p []byte, v any) error {
	rootCfg := configuration.ParseString(string(p))
	result := make(map[string]any)

	for _, key := range rootCfg.Root().GetObject().GetKeys() {
		result[key] = getValue(rootCfg, key)
	}

	return convert.Remarshal("hocon", result, v)
}

// getValue returns the value bound at the given path. Only paths that hold an
// object are traversed, so the object of a scalar is never asked for its keys.
func getValue(rootCfg *configuration.Config, path string) any {
	if rootCfg.IsObject(path) {
		return getConfig(rootCfg, rootCfg.GetConfig(path), path)
	}

	return convertType(rootCfg.GetValue(path))
}

func getConfig(rootCfg, cfg *configuration.Config, path string) map[string]any {
	result := make(map[string]any)

	for _, key := range cfg.Root().GetObject().GetKeys() {
		result[key] = getValue(rootCfg, path+"."+key)
	}

	return result
}

func convertType(value *hocon.HoconValue) any {
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
