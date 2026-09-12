package ini

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-ini/ini"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Parser is an INI parser.
type Parser struct{}

// Unmarshal unmarshals INI files.
func (i *Parser) Unmarshal(p []byte, v any) error {
	cfg, err := ini.Load(p)
	if err != nil {
		return fmt.Errorf("read ini file: %w", err)
	}

	result := make(map[string]map[string]any)
	for _, s := range cfg.Sections() {
		sectionName := s.Name()
		if sectionName == "DEFAULT" {
			continue
		}

		result[sectionName] = map[string]any{}
		keysHash := s.KeysHash()
		result[sectionName] = convertKeyTypes(keysHash)
	}

	return convert.Remarshal("ini", result, v)
}

func convertKeyTypes(keysHash map[string]string) map[string]any {
	val := map[string]any{}

	for k, v := range keysHash {
		switch {
		case isNumberLiteral(v):
			f, _ := strconv.ParseFloat(v, 64)
			val[k] = f
		case isBooleanLiteral(v):
			b, _ := strconv.ParseBool(v)
			val[k] = b
		default:
			val[k] = v
		}
	}

	return val
}

func isNumberLiteral(f string) bool {
	n, err := strconv.ParseFloat(f, 64)
	return err == nil && !math.IsNaN(n) && !math.IsInf(n, 0)
}

func isBooleanLiteral(b string) bool {
	_, err := strconv.ParseBool(b)
	return err == nil
}
