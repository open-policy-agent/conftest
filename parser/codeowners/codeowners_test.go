package codeowners

import (
	"encoding/json"
	"fmt"
	"testing"
)

const appsGitHubPath = "/apps/github"

var (
	bobEmail       = ownerFromEmail("bob@example.test")
	bobUser        = ownerFromUsername("bob")
	carolUser      = ownerFromUsername("carol")
	docsEmail      = ownerFromEmail("docs@example.com")
	doctocatUser   = ownerFromUsername("doctocat")
	octocatUser    = ownerFromUsername("octocat")
	developersTeam = ownerFromTeam("my-org/developers")
	librariansTeam = ownerFromTeam("my-org/librarians")
	octocatTeam    = ownerFromTeam("octo-org/octocats")
	globalOwner1   = ownerFromUsername("global-owner1")
	globalOwner2   = ownerFromUsername("global-owner2")
	jsOwner        = ownerFromUsername("js-owner")
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
			input: []byte(`/apps/github`),
			want: owners{
				Rules: []rule{
					{
						Pattern: appsGitHubPath,
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
			name:  "one character wildcard",
			input: []byte(`b?in/ @my-org/developers`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "b?in/",
						Owners:  []owner{developersTeam},
					},
				},
			},
		},
		{
			name:  "one character wildcard",
			input: []byte(`b?in/ @my-org/developers`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "b?in/",
						Owners:  []owner{developersTeam},
					},
				},
			},
		},
		{
			name:  "literal asterisk",
			input: []byte(`a\*b @my-org/developers`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "a\\*b",
						Owners:  []owner{developersTeam},
					},
				},
			},
		},
		{
			name:  "literal question mark",
			input: []byte(`a\?b @my-org/developers`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "a\\?b",
						Owners:  []owner{developersTeam},
					},
				},
			},
		},
		// GitHub does not support character ranges (e.g., [a-z])
		{
			name:  "recursive match",
			input: []byte(`**/test @my-org/developers`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "**/test",
						Owners:  []owner{developersTeam},
					},
				},
			},
		},
		{
			name: "kitchen sink",
			// Example from GitHub's docs
			// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
			input: []byte(`# This is a comment.
# Each line is a file pattern followed by one or more owners.

# These owners will be the default owners for everything in
# the repo. Unless a later match takes precedence,
# @global-owner1 and @global-owner2 will be requested for
# review when someone opens a pull request.
*       @global-owner1 @global-owner2

# Order is important; the last matching pattern takes the most
# precedence. When someone opens a pull request that only
# modifies JS files, only @js-owner and not the global
# owner(s) will be requested for a review.
*.js    @js-owner #This is an inline comment.

# You can also use email addresses if you prefer. They'll be
# used to look up users just like we do for commit author
# emails.
*.go docs@example.com

# Teams can be specified as code owners as well. Teams should
# be identified in the format @org/team-name. Teams must have
# explicit write access to the repository. In this example,
# the octocats team in the octo-org organization owns all .txt files.
*.txt @octo-org/octocats

# In this example, @doctocat owns any files in the build/logs
# directory at the root of the repository and any of its
# subdirectories.
/build/logs/ @doctocat

# The "docs/*"" pattern will match files like
# "docs/getting-started.md" but not further nested files like
# docs/build-app/troubleshooting.md.
docs/* docs@example.com

# In this example, @octocat owns any file in an apps directory
# anywhere in your repository.
apps/ @octocat

# In this example, @doctocat owns any file in the "/docs"
# directory in the root of your repository and any of its
# subdirectories.
/docs/ @doctocat

# In this example, any change inside the "/scripts" directory
# will require approval from @doctocat or @octocat.
/scripts/ @doctocat @octocat

# In this example, @octocat owns any file in a "/logs" directory such as
# "/build/logs", "/scripts/logs", and "/deeply/nested/logs". Any changes
# in a "/logs" directory will require approval from @octocat.
**/logs @octocat

# In this example, @octocat owns any file in the "/apps"
# directory in the root of your repository except for the appsGitHubPath
# subdirectory, as its owners are left empty. Without an owner, changes
# to "apps/github" can be made with the approval of any user who has
# write access to the repository.
/apps/ @octocat
/apps/github

# In this example, @octocat owns any file in the "/apps"
# directory in the root of your repository except for the appsGitHubPath
# subdirectory, as this subdirectory has its own owner @doctocat
/apps/ @octocat
/apps/github @doctocat
`),
			want: owners{
				Rules: []rule{
					{
						Pattern: "*",
						Owners:  []owner{globalOwner1, globalOwner2},
					},
					{
						Pattern: "*.js",
						Owners:  []owner{jsOwner},
					},
					{
						Pattern: "*.go",
						Owners:  []owner{docsEmail},
					},
					{
						Pattern: "*.txt",
						Owners:  []owner{octocatTeam},
					},
					{
						Pattern: "/build/logs/",
						Owners:  []owner{doctocatUser},
					},
					{
						Pattern: "docs/*",
						Owners:  []owner{docsEmail},
					},
					{
						Pattern: "apps/",
						Owners:  []owner{octocatUser},
					},
					{
						Pattern: "/docs/",
						Owners:  []owner{doctocatUser},
					},
					{
						Pattern: "/scripts/",
						Owners:  []owner{doctocatUser, octocatUser},
					},
					{
						Pattern: "**/logs",
						Owners:  []owner{octocatUser},
					},
					{
						Pattern: "/apps/",
						Owners:  []owner{octocatUser},
					},
					{
						Pattern: appsGitHubPath,
						Owners:  []owner{},
					},
					{
						Pattern: "/apps/",
						Owners:  []owner{octocatUser},
					},
					{
						Pattern: appsGitHubPath,
						Owners:  []owner{doctocatUser},
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
