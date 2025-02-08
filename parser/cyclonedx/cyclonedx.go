package cyclonedx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/CycloneDX/cyclonedx-go"
)

// Parser is a CycloneDX parser.
type Parser struct{}

// Unmarshal unmarshals CycloneDX files.
func (*Parser) Unmarshal(p []byte, v any) error {
	bomFileFormat := cyclonedx.BOMFileFormatJSON
	if !json.Valid(p) {
		bomFileFormat = cyclonedx.BOMFileFormatXML
	}
	bom := new(cyclonedx.BOM)
	decoder := cyclonedx.NewBOMDecoder(bytes.NewBuffer(p), bomFileFormat)
	if err := decoder.Decode(bom); err != nil {
		panic(err)
	}

	temp := p

	if bomFileFormat == cyclonedx.BOMFileFormatXML {
		var data cyclonedx.BOM
		if err := xml.Unmarshal(p, &data); err != nil {
			return fmt.Errorf("unmarshaling XML error: %v", err)
		}
		if d, err := json.Marshal(data); err == nil {
			temp = d
		} else {
			return fmt.Errorf("marshaling JSON error: %v", err)
		}
	}

	err := json.Unmarshal(temp, v)
	if err != nil {
		return fmt.Errorf("unmarshaling JSON error: %v", err)
	}

	return nil
}
