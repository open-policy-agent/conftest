package codeowners

import (
	"encoding/json"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    codeowners
		wantErr bool
	}{
		{
			name: "anyone with write access",
			// A lone pattern requires a review from anyone with write access
			input: []byte(`/apps/github`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "/apps/github",
						Owners:  []string{},
					},
				},
			},
		},
		{
			name:  "comment",
			input: []byte(`# Just a comment`),
			want: codeowners{
				Entries: []entry{},
			},
		},
		{
			name:  "user owner",
			input: []byte(`* @bob`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*",
						Owners:  []string{"@bob"},
					},
				},
			},
		},
		{
			name:  "email owner",
			input: []byte(`*.go bob@example.test`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*.go",
						Owners:  []string{"bob@example.test"},
					},
				},
			},
		},
		{
			name:  "team owner",
			input: []byte(`*.txt @my-org/developers`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*.txt",
						Owners:  []string{"@my-org/developers"},
					},
				},
			},
		},
		{
			name:  "two owners",
			input: []byte(`* @bob @carol`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*",
						Owners:  []string{"@bob", "@carol"},
					},
				},
			},
		},
		{
			name:  "inline comment",
			input: []byte(`*.js    @carol #This is an inline comment.`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*.js",
						Owners:  []string{"@carol"},
					},
				},
			},
		},
		{
			name:  "folder path",
			input: []byte(`/build/logs/ @my-org/librarians`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "/build/logs/",
						Owners:  []string{"@my-org/librarians"},
					},
				},
			},
		},
		{
			name:  "one character wildcard",
			input: []byte(`b?in/ @my-org/developers`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "b?in/",
						Owners:  []string{"@my-org/developers"},
					},
				},
			},
		},
		{
			name:  "one character wildcard",
			input: []byte(`b?in/ @my-org/developers`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "b?in/",
						Owners:  []string{"@my-org/developers"},
					},
				},
			},
		},
		{
			name:  "literal asterisk",
			input: []byte(`a\*b @my-org/developers`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "a\\*b",
						Owners:  []string{"@my-org/developers"},
					},
				},
			},
		},
		{
			name:  "literal question mark",
			input: []byte(`a\?b @my-org/developers`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "a\\?b",
						Owners:  []string{"@my-org/developers"},
					},
				},
			},
		},
		// GitHub does not support character ranges (e.g., [a-z])
		{
			name:  "recursive match",
			input: []byte(`**/test @my-org/developers`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "**/test",
						Owners:  []string{"@my-org/developers"},
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
# directory in the root of your repository except for the "/apps/github"
# subdirectory, as its owners are left empty. Without an owner, changes
# to "apps/github" can be made with the approval of any user who has
# write access to the repository.
/apps/ @octocat
/apps/github

# In this example, @octocat owns any file in the "/apps"
# directory in the root of your repository except for the "/apps/github"
# subdirectory, as this subdirectory has its own owner @doctocat
/apps/ @octocat
/apps/github @doctocat
`),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*",
						Owners:  []string{"@global-owner1", "@global-owner2"},
					},
					{
						Pattern: "*.js",
						Owners:  []string{"@js-owner"},
					},
					{
						Pattern: "*.go",
						Owners:  []string{"docs@example.com"},
					},
					{
						Pattern: "*.txt",
						Owners:  []string{"@octo-org/octocats"},
					},
					{
						Pattern: "/build/logs/",
						Owners:  []string{"@doctocat"},
					},
					{
						Pattern: "docs/*",
						Owners:  []string{"docs@example.com"},
					},
					{
						Pattern: "apps/",
						Owners:  []string{"@octocat"},
					},
					{
						Pattern: "/docs/",
						Owners:  []string{"@doctocat"},
					},
					{
						Pattern: "/scripts/",
						Owners:  []string{"@doctocat", "@octocat"},
					},
					{
						Pattern: "**/logs",
						Owners:  []string{"@octocat"},
					},
					{
						Pattern: "/apps/",
						Owners:  []string{"@octocat"},
					},
					{
						Pattern: "/apps/github",
						Owners:  []string{},
					},
					{
						Pattern: "/apps/",
						Owners:  []string{"@octocat"},
					},
					{
						Pattern: "/apps/github",
						Owners:  []string{"@doctocat"},
					},
				},
			},
		},
		{
			name:  "windows line endings",
			input: []byte("# developers by default\r\n* @my-org/developers\r\n# librarians for docs\r\ndocs/\t@my-org/librarians\r\n"),
			want: codeowners{
				Entries: []entry{
					{
						Pattern: "*",
						Owners:  []string{"@my-org/developers"},
					},
					{
						Pattern: "docs/",
						Owners:  []string{"@my-org/librarians"},
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

func roundTrip(t *testing.T, parsed any) codeowners {
	t.Helper()
	codeownersJSON, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("marshal CODEOWNERS to JSON: %v", err)
	}
	var output codeowners
	if err := json.Unmarshal(codeownersJSON, &output); err != nil {
		t.Fatalf("unmarshal CODEOWNERS JSON: %v", err)
	}
	return output
}

func expectCodeowners(t *testing.T, got, want codeowners) {
	t.Helper()
	if len(got.Entries) != len(want.Entries) {
		t.Errorf("expected %d entries, got %d entries", len(want.Entries), len(got.Entries))
		return
	}
	for i, wantEntry := range want.Entries {
		gotEntry := got.Entries[i]
		expectEntry(t, i, gotEntry, wantEntry)
	}
}

func expectEntry(t *testing.T, entryNumber int, got, want entry) {
	if got.Pattern != want.Pattern {
		t.Errorf("entry %d: expected '%s' pattern, got '%s'", entryNumber, want.Pattern, got.Pattern)
	}
	if len(got.Owners) != len(want.Owners) {
		t.Errorf("entry %d: expected %d owners, got %d owners", entryNumber, len(want.Owners), len(got.Owners))
		return
	}
	for i, wantOwner := range want.Owners {
		gotOwner := got.Owners[i]
		if gotOwner != wantOwner {
			t.Errorf("entry %d: owner %d: expected '%s', got '%s'", entryNumber, i, wantOwner, gotOwner)
		}
	}
}
