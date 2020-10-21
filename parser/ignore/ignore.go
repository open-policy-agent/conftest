package ignore

import (
	"encoding/json"

	ignore "github.com/shteou/go-ignore"
)

// Parser is a ignore (dockerignore, gitignore) parser.
type Parser struct{}

// Unmarshal unmarshals ignore files.
func (pp *Parser) Unmarshal(p []byte, v interface{}) error {
	ignoreEntries, err := ignore.ParseIgnoreBytes(p)
	if err != nil {
		return err
	}

	// Wrap the entry list in another list, to ensure it's
	// treated as a single file
	entryListList := [][]ignore.Entry{ignoreEntries}

	marshalledLines, err := json.Marshal(entryListList)
	if err != nil {
		return err
	}

	json.Unmarshal(marshalledLines, v)
	return nil
}
