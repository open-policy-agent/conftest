// Package routeros parses MikroTik RouterOS configuration exports, the output
// produced by the RouterOS "/export" command.
package routeros

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Parser is a RouterOS configuration parser.
type Parser struct{}

// entry is a single command within a configuration section, for example an
// "add" or "set" line. The command verb is stored under "command" and any
// tokens that are not key=value pairs (selectors like "[ find ]", positional
// ids, or "!"-prefixed flags) are stored under "arguments".
type entry map[string]any

// Unmarshal unmarshals a RouterOS configuration export. The result is an object
// keyed by section path (for example "/ip address"), where each value is the
// list of commands declared under that section in the order they appear.
func (p *Parser) Unmarshal(input []byte, out any) error {
	sections := map[string][]entry{}
	var order []string

	var current string
	haveSection := false

	for _, line := range logicalLines(string(input)) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "/") {
			current = trimmed
			haveSection = true
			if _, ok := sections[current]; !ok {
				sections[current] = []entry{}
				order = append(order, current)
			}
			continue
		}

		if !haveSection {
			return fmt.Errorf("command %q found before any section header", trimmed)
		}

		sections[current] = append(sections[current], parseCommand(trimmed))
	}

	// Preserve section order so the JSON round-trip is deterministic.
	result := make(map[string]any, len(sections))
	for _, name := range order {
		result[name] = sections[name]
	}

	j, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal routeros to json: %w", err)
	}
	if err := json.Unmarshal(j, out); err != nil {
		return fmt.Errorf("unmarshal routeros json: %w", err)
	}
	return nil
}

// logicalLines joins RouterOS line continuations. The exporter wraps long lines
// with a trailing backslash, then indents the continuation. Joining strips the
// backslash and the continuation's leading whitespace so wrapped values are
// reassembled exactly as RouterOS stored them.
func logicalLines(s string) []string {
	physical := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")

	var out []string
	var buf strings.Builder
	continuing := false

	for _, line := range physical {
		part := line
		if continuing {
			part = strings.TrimLeft(part, " \t")
		}

		if strings.HasSuffix(part, "\\") {
			buf.WriteString(strings.TrimSuffix(part, "\\"))
			continuing = true
			continue
		}

		buf.WriteString(part)
		out = append(out, buf.String())
		buf.Reset()
		continuing = false
	}

	if continuing {
		out = append(out, buf.String())
	}
	return out
}

// parseCommand splits a command line into its verb, key=value pairs, and any
// remaining positional arguments or flags.
func parseCommand(line string) entry {
	tokens := tokenize(line)
	e := entry{}
	var args []string

	for i, tok := range tokens {
		if i == 0 {
			e["command"] = tok
			continue
		}

		// Bracketed selectors and "!" flags are not assignments, even though a
		// selector such as "[ find name=all ]" contains an "=".
		if strings.HasPrefix(tok, "[") || strings.HasPrefix(tok, "!") {
			args = append(args, tok)
			continue
		}

		if key, value, ok := strings.Cut(tok, "="); ok {
			e[key] = unquote(value)
			continue
		}

		args = append(args, tok)
	}

	if len(args) > 0 {
		e["arguments"] = args
	}
	return e
}

// tokenize splits a logical line on whitespace while keeping double-quoted
// strings and bracketed "[ ... ]" expressions as single tokens.
func tokenize(line string) []string {
	var tokens []string
	var cur strings.Builder
	inQuote := false
	escaped := false
	depth := 0

	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}

	for _, r := range line {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			cur.WriteRune(r)
			escaped = true
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case inQuote:
			cur.WriteRune(r)
		case r == '[':
			depth++
			cur.WriteRune(r)
		case r == ']':
			if depth > 0 {
				depth--
			}
			cur.WriteRune(r)
		case (r == ' ' || r == '\t') && depth == 0:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return tokens
}

// unquote removes surrounding double quotes and unescapes the escape sequences
// RouterOS emits inside quoted values. Unquoted values are returned unchanged.
func unquote(value string) string {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return value
	}

	inner := value[1 : len(value)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' && i+1 < len(inner) {
			next := inner[i+1]
			switch next {
			case '"', '\\':
				b.WriteByte(next)
				i++
				continue
			}
		}
		b.WriteByte(inner[i])
	}
	return b.String()
}
