package ignore

import (
	"fmt"
	"strings"

	ignore "github.com/shteou/go-ignore"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Parser is a ignore (dockerignore, gitignore) parser.
type Parser struct{}

// Unmarshal unmarshals ignore files.
func (pp *Parser) Unmarshal(p []byte, v any) error {
	ignoreEntries, err := parseLines(strings.Split(string(p), "\n"))
	if err != nil {
		return fmt.Errorf("parse ignore bytes: %w", err)
	}

	// Wrap the entry list in another list, to ensure it's
	// treated as a single file.
	entryListList := [][]ignore.Entry{ignoreEntries}

	return convert.Remarshal("ignore", entryListList, v)
}

// danglingEscapeGuard is appended to lines that end in a backslash before they
// are handed to go-ignore. Its TrimRightSpace slices past the end of such a
// line and panics. The guard gives the backslash something to escape, and is
// removed again from the result.
const danglingEscapeGuard = "x"

func parseLines(lines []string) ([]ignore.Entry, error) {
	guarded := make([]string, len(lines))
	for i, l := range lines {
		guarded[i] = l
		if strings.HasSuffix(l, `\`) {
			guarded[i] = l + danglingEscapeGuard
		}
	}

	entries, err := ignore.ParseIgnoreLines(guarded)
	if err != nil {
		return nil, err
	}

	for i := range entries {
		if guarded[i] == lines[i] {
			continue
		}
		entries[i].Value = strings.TrimSuffix(entries[i].Value, danglingEscapeGuard)
		entries[i].Original = lines[i]
	}

	return entries, nil
}
