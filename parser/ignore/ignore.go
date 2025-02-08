package ignore

import (
	"encoding/json"
	"fmt"

	ignore "github.com/shteou/go-ignore"
)

// Parser is a ignore (dockerignore, gitignore) parser.
type Parser struct{}

// Unmarshal unmarshals ignore files.
func (pp *Parser) Unmarshal(p []byte, v any) error {
	ignoreEntries, err := ignore.ParseIgnoreBytes(p)
	if err != nil {
		return fmt.Errorf("parse ignore bytes: %w", err)
	}

	// Wrap the entry list in another list, to ensure it's
	// treated as a single file.
	entryListList := [][]ignore.Entry{ignoreEntries}

	marshalledLines, err := json.Marshal(entryListList)
	if err != nil {
		return fmt.Errorf("marshal ignore: %w", err)
	}

	if err := json.Unmarshal(marshalledLines, v); err != nil {
		return fmt.Errorf("unmarshal ignore: %w", err)
	}

	return nil
}
