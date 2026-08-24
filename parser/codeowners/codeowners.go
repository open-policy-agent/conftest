// Package codeowners provides a parser for CODEOWNERS files
//
// A CODEOWNERS file defines which reviewers are required for various
// files within a repository.
//
// This parser is for the [format used by GitHub]. The format is
// similar to the [gitignore] format, but lacks the following features:
//   - The comment character '#' cannot be escaped
//   - Patterns cannot be negated with '!'
//   - Character ranges, like `[a-z]` are not supported
//
// GitHub searches for the CODEOWNERS file in the following priority
// order relative to the repository root:
//   - .github/CODEOWNERS
//   - CODEOWNERS
//   - docs/CODEOWNERS
//
// [format used by GitHub]: https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
// [gitignore]: https://git-scm.com/docs/gitignore
package codeowners

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Token patterns for a CODEOWNERS file.
var (
	comment    = regexp.MustCompile(`^#(.*)((\r?\n)|\z)`)
	whitespace = regexp.MustCompile(`^[\t\f ]+`)
	pattern    = regexp.MustCompile(`^[\w/\\*?.-]+`)
	owner      = regexp.MustCompile(`^[\w@/.-]+`)
)

// codeowners contains required reviewers for zero or more file patterns.
type codeowners struct {
	Entries []entry `json:"entries"`
}

// entry contains the required reviewers for a file pattern.
type entry struct {
	Pattern string   `json:"pattern"`
	Owners  []string `json:"owners"`
}

// state stores the progress for a parsing pass.
type state struct {
	data   []byte
	line   int
	column int
}

// newState initializes the parser state for a parsing pass.
func newState(data []byte) *state {
	return &state{
		data: data,
		line: 1,
	}
}

// parse attempts to parse the CODEOWNERS data.
func (s *state) parse() (*codeowners, error) {
	var entries []entry
	for !s.done() {
		// Trim any leading whitespace
		s.matchWhitespace()

		// Empty line
		if s.matchNewline() {
			continue
		}

		// Comment line
		if s.matchComment() {
			continue
		}

		// Line with an entry
		entry, err := s.matchEntry()
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return &codeowners{
		Entries: entries,
	}, nil
}

// done checks if the parser is at the end of the data.
func (s *state) done() bool {
	return len(s.data) == 0
}

// matchWhitespace consumes whitespace characters other than newlines.
// It returns if any whitespace characters were consumed.
func (s *state) matchWhitespace() bool {
	loc := whitespace.FindIndex(s.data)
	if loc == nil {
		return false
	}
	s.data = s.data[loc[1]:]
	s.column += loc[1]
	return true
}

// matchNewline consumes a newline. It returns true if a newline was
// consumed or if the parser is at the end of the data.
func (s *state) matchNewline() bool {
	// End of data
	if s.done() {
		return true
	}
	// Windows line endings
	if len(s.data) >= 2 && s.data[0] == '\r' && s.data[1] == '\n' {
		s.data = s.data[2:]
		s.column = 0
		s.line += 1
		return true
	}
	// Unix line endings
	if s.data[0] == '\n' {
		s.data = s.data[1:]
		s.column = 0
		s.line += 1
		return true
	}
	return false
}

// matchComment consumes a comment. It returns true if a comment was
// consumed.
func (s *state) matchComment() bool {
	loc := comment.FindIndex(s.data)
	if loc == nil {
		return false
	}
	s.data = s.data[loc[1]:]
	s.column += loc[1]
	return true
}

// matchEntry parses the ownership for a single file pattern. It
// returns an entry for the pattern or an error if there were issues
// parsing the entry.
func (s *state) matchEntry() (entry, error) {
	output := entry{}
	var err error
	// Entry starts with a pattern
	output.Pattern, err = s.matchPattern()
	if err != nil {
		return output, err
	}
	// Pattern is followed by zero or more owners separated by whitespace
	for {
		hadSpace := s.matchWhitespace()
		if s.matchNewline() {
			break
		}
		if s.matchComment() {
			break
		}
		if !hadSpace {
			return output, fmt.Errorf("%d:%d: invalid owner separation", s.line, s.column)
		}
		owner, err := s.matchOwner()
		if err != nil {
			return output, err
		}
		output.Owners = append(output.Owners, owner)
	}
	return output, nil
}

// matchPattern parses a file pattern.
func (s *state) matchPattern() (string, error) {
	loc := pattern.FindIndex(s.data)
	if loc == nil {
		return "", fmt.Errorf("%d:%d: expected pattern", s.line, s.column)
	}
	output := string(s.data[:loc[1]])
	s.data = s.data[loc[1]:]
	s.column += loc[1]
	return output, nil
}

// matchOwner parses an owner, which can be a username, team, or email address.
func (s *state) matchOwner() (string, error) {
	loc := owner.FindIndex(s.data)
	if loc == nil {
		return "", fmt.Errorf("%d:%d: expected owner", s.line, s.column)
	}
	output := string(s.data[loc[0]:loc[1]])
	s.data = s.data[loc[1]:]
	s.column += loc[1]
	return output, nil
}

// Parser is a CODEOWNERS file parser.
type Parser struct{}

// Unmarshal unmarshals a CODEOWNERS file.
func (p *Parser) Unmarshal(data []byte, v any) error {
	s := newState(data)
	config, err := s.parse()
	if err != nil {
		return err
	}

	j, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal CODEOWNERS to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal CODEOWNERS json: %w", err)
	}

	return nil
}
