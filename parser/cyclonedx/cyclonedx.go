package cyclonedx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/CycloneDX/cyclonedx-go"
)

// Parser is a CycloneDX parser.
type Parser struct{}

// Parse parses CycloneDX files.
func (*Parser) Parse(r io.Reader) ([]any, error) {
	p, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
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
			return nil, fmt.Errorf("unmarshaling XML error: %v", err)
		}
		if d, err := json.Marshal(data); err == nil {
			temp = d
		} else {
			return nil, fmt.Errorf("marshaling JSON error: %v", err)
		}
	}

	var v any
	if err := json.Unmarshal(temp, &v); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON error: %v", err)
	}

	return []any{v}, nil
}
