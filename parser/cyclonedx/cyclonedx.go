package cyclonedx

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/CycloneDX/cyclonedx-go"
)

// Parser is a CycloneDX parser.
type Parser struct{}

// Unmarshal unmarshals CycloneDX files.
func (*Parser) Unmarshal(p []byte, v interface{}) error {
	bom := new(cyclonedx.BOM)
	decoder := cyclonedx.NewBOMDecoder(bytes.NewBuffer(p), cyclonedx.BOMFileFormatJSON)
	if err := decoder.Decode(bom); err != nil {
		panic(err)
	}

	err := json.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("unmarshalling error: %v", err)
	}

	return nil
}
