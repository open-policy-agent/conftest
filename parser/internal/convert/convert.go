// Package convert provides conversion helpers shared by the parsers.
package convert

import (
	"encoding/json"
	"fmt"
)

// Remarshal marshals in to JSON and unmarshals the result into out, which
// normalizes a parser's intermediate value into plain maps, slices and
// scalars. The name identifies the parser in error messages.
func Remarshal(name string, in, out any) error {
	j, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal %s to json: %w", name, err)
	}

	if err := json.Unmarshal(j, out); err != nil {
		return fmt.Errorf("unmarshal %s json: %w", name, err)
	}

	return nil
}
