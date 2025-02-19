package edn

import (
	"fmt"

	"olympos.io/encoding/edn"
)

// Parser is an EDN parser.
type Parser struct{}

// Unmarshal unmarshals EDN encoded files.
func (tp *Parser) Unmarshal(p []byte, v any) error {
	var res any

	if err := edn.Unmarshal(p, &res); err != nil {
		return fmt.Errorf("unmarshal EDN: %w", err)
	}

	*v.(*any) = cleanupMapValue(res)

	return nil
}

func cleanupInterfaceArray(in []any) []any {
	res := make([]any, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceMap(in map[any]any) map[string]any {
	res := make(map[string]any)
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v any) any {
	switch v := v.(type) {
	case []any:
		return cleanupInterfaceArray(v)
	case map[any]any:
		return cleanupInterfaceMap(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
