package codeowners

import (
	"encoding/json"
	"fmt"
	"testing"
)

const codePath = "/apps/code.go"

var (
	bobEmail       = ownerFromEmail("bob@example.test")
	bobUser        = ownerFromUsername("bob")
	carolUser      = ownerFromUsername("carol")
	developersTeam = ownerFromTeam("my-org/developers")
	librariansTeam = ownerFromTeam("my-org/librarians")
	goOwner        = ownerFromUsername("go-owner")
)

func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    owners
		wantErr bool
	}{
		{
			name: "anyone with write access",
			// A lone pattern requires a review from anyone with write access
			input: []byte(codePath),
			want: owners{
				Rules: []rule{
					{
						Pattern: codePath,
						Owners:  []owner{},
					},
				},
			},
		},
		{
			name:  "comment",
			input: []byte(`# Just a comment`),
			want: owners{
				Rules: []rule{},
			},
		},
		{
			name:  "user owner",
			input: []byte(`* @bob`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*",
						Owners:  []owner{bobUser},
					},
				},
			},
		},
		{
			name:  "email owner",
			input: []byte(`*.go bob@example.test`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*.go",
						Owners:  []owner{bobEmail},
					},
				},
			},
		},
		{
			name:  "team owner",
			input: []byte(`*.txt @my-org/developers`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*.txt",
						Owners:  []owner{developersTeam},
					},
				},
			},
		},
		{
			name:  "two owners",
			input: []byte(`* @bob @carol`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*",
						Owners:  []owner{bobUser, carolUser},
					},
				},
			},
		},
		{
			name:  "inline comment",
			input: []byte(`*.js    @carol #This is an inline comment.`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*.js",
						Owners:  []owner{carolUser},
					},
				},
			},
		},
		{
			name:  "folder path",
			input: []byte(`/build/logs/ @my-org/librarians`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "/build/logs/",
						Owners:  []owner{librariansTeam},
					},
				},
			},
		},
		{
			name: "multi-line example",
			input: []byte(`# This is a comment.
# Default owner
*       @my-org/developers

# Special owner for go files
*.go @go-owner

*.txt @bob

docs/ @my-org/librarians

/scripts/ @bob @carol
`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*",
						Owners:  []owner{developersTeam},
					},
					{
						Pattern: "*.go",
						Owners:  []owner{goOwner},
					},
					{
						Pattern: "*.txt",
						Owners:  []owner{bobUser},
					},
					{
						Pattern: "docs/",
						Owners:  []owner{librariansTeam},
					},
					{
						Pattern: "/scripts/",
						Owners:  []owner{bobUser, carolUser},
					},
				},
			},
		},
		{
			name:  "windows line endings",
			input: []byte("# developers by default\r\n* @my-org/developers\r\n# librarians for docs\r\ndocs/\t@my-org/librarians\r\n"),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*",
						Owners:  []owner{developersTeam},
					},
					{
						Pattern: "docs/",
						Owners:  []owner{librariansTeam},
					},
				},
			},
		},
	}

	parser := &Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsedInput any
			err := parser.Unmarshal(tt.input, &parsedInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && parsedInput == nil {
				t.Fatal("expected parsed content, got nil")
			}

			got := roundTrip(t, parsedInput)
			if err != nil {
				t.Fatalf("round trip of CODEOWNERS failed: %v", err)
			}
			expectCodeowners(t, got, tt.want)
		})
	}
}

func roundTrip(t *testing.T, parsed any) owners {
	t.Helper()
	ownersJSON, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("marshal CODEOWNERS to JSON: %v", err)
	}
	var output owners
	if err := json.Unmarshal(ownersJSON, &output); err != nil {
		t.Fatalf("unmarshal CODEOWNERS JSON: %v", err)
	}
	return output
}

func expectCodeowners(t *testing.T, got, want owners) {
	t.Helper()
	if len(got.Rules) != len(want.Rules) {
		t.Errorf("expected %d rules, got %d rules", len(want.Rules), len(got.Rules))
		return
	}
	for i, wantRule := range want.Rules {
		gotRule := got.Rules[i]
		expectRule(t, gotRule, wantRule)
	}
}

func expectRule(t *testing.T, got, want rule) {
	t.Helper()
	if got.Pattern != want.Pattern {
		t.Errorf("line %d: expected '%s' pattern, got '%s'", got.LineNumber, want.Pattern, got.Pattern)
	}
	if len(got.Owners) != len(want.Owners) {
		t.Errorf("line %d: expected %d owners, got %d owners", got.LineNumber, len(want.Owners), len(got.Owners))
		return
	}
	for i, wantOwner := range want.Owners {
		gotOwner := got.Owners[i]
		err := expectOwner(gotOwner, wantOwner)
		if err != nil {
			t.Errorf("line %d: owner %d: %v", got.LineNumber, i, err)
		}
	}
}

func expectOwner(got, want owner) error {
	if got.Type != want.Type || got.Value != want.Value {
		return fmt.Errorf("expected: %s %s, got %s %s", want.Type, want.Value, got.Type, got.Value)
	}
	return nil
}
