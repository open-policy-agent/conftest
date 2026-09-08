package ignore

import (
	"fmt"

	ignore "github.com/shteou/go-ignore"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
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

	return convert.Remarshal("ignore", entryListList, v)
}
