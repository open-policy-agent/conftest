// Package codeowners provides a parser for CODEOWNERS files
//
// A [CODEOWNERS] file defines which reviewers are required for various
// files within a repository. The file format is similar to
// [gitignore].
//
// GitHub searches for the CODEOWNERS file in the following priority
// order relative to the repository root:
//   - .github/CODEOWNERS
//   - CODEOWNERS
//   - docs/CODEOWNERS
//
// [CODEOWNERS]: https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
// [gitignore]: https://git-scm.com/docs/gitignore
package codeowners

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hmarr/codeowners"
)

// codeowners contains required reviewers for zero or more file patterns.
type owners struct {
	Rules []rule `json:"rules"`
}

// rule contains the required reviewers for a file pattern.
type rule struct {
	Pattern    string  `json:"pattern"`
	Owners     []owner `json:"owners"`
	Comment    string  `json:"comment"`
	LineNumber int     `json:"lineNumber"`
}

type ownerType string

const (
	// EmailOwner is the owner type for email addresses.
	emailOwner ownerType = "email"
	// TeamOwner is the owner type for GitHub teams.
	teamOwner ownerType = "team"
	// UsernameOwner is the owner type for GitHub usernames.
	usernameOwner ownerType = "username"
)

// owner contains information about a codeowner
type owner struct {
	Type  ownerType `json:"type"`
	Value string    `json:"value"`
}

func ownerFromEmail(email string) owner {
	return owner{
		Type:  emailOwner,
		Value: email,
	}
}

func ownerFromTeam(team string) owner {
	return owner{
		Type:  teamOwner,
		Value: team,
	}
}

func ownerFromUsername(username string) owner {
	return owner{
		Type:  usernameOwner,
		Value: username,
	}
}

// Parser is a CODEOWNERS file parser.
type Parser struct{}

// Unmarshal unmarshals a CODEOWNERS file.
func (p *Parser) Unmarshal(data []byte, v any) error {
	f := bytes.NewBuffer(data)
	ruleset, err := codeowners.ParseFile(f)
	if err != nil {
		return fmt.Errorf("failed to parse CODEOWNERS: %w", err)
	}

	config := owners{}
	for _, r := range ruleset {
		os := make([]owner, len(r.Owners))
		for i, o := range r.Owners {
			os[i] = owner{
				Type:  ownerType(o.Type),
				Value: o.Value,
			}
		}
		config.Rules = append(config.Rules, rule{
			Pattern:    r.RawPattern(),
			Owners:     os,
			Comment:    r.Comment,
			LineNumber: r.LineNumber,
		})
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
