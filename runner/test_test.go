package runner

import (
	"reflect"
	"testing"
)

func TestHasWildcard(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{
			name:     "no wildcard",
			patterns: []string{"main", "test"},
			want:     false,
		},
		{
			name:     "asterisk wildcard",
			patterns: []string{"main.*"},
			want:     true,
		},
		{
			name:     "question mark wildcard",
			patterns: []string{"main.?"},
			want:     true,
		},
		{
			name:     "bracket wildcard",
			patterns: []string{"main.[abc]"},
			want:     true,
		},
		{
			name:     "mixed patterns",
			patterns: []string{"main", "test.*"},
			want:     true,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasWildcard(tt.patterns); got != tt.want {
				t.Errorf("hasWildcard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterNamespaces(t *testing.T) {
	tests := []struct {
		name      string
		available []string
		patterns  []string
		want      []string
	}{
		{
			name:      "asterisk matches suffix",
			available: []string{"main", "main.gke", "main.aws", "test"},
			patterns:  []string{"main.*"},
			want:      []string{"main.gke", "main.aws"},
		},
		{
			name:      "asterisk matches prefix",
			available: []string{"main.gke", "test.gke", "main.aws"},
			patterns:  []string{"*.gke"},
			want:      []string{"main.gke", "test.gke"},
		},
		{
			name:      "exact match without wildcard",
			available: []string{"main", "main.gke"},
			patterns:  []string{"main"},
			want:      []string{"main"},
		},
		{
			name:      "multiple patterns",
			available: []string{"main", "main.gke", "test", "test.aws"},
			patterns:  []string{"main.*", "test.*"},
			want:      []string{"main.gke", "test.aws"},
		},
		{
			name:      "no matches",
			available: []string{"main", "test"},
			patterns:  []string{"foo.*"},
			want:      nil,
		},
		{
			name:      "match all with star",
			available: []string{"main", "test", "foo"},
			patterns:  []string{"*"},
			want:      []string{"main", "test", "foo"},
		},
		{
			name:      "question mark pattern",
			available: []string{"main.a", "main.b", "main.ab"},
			patterns:  []string{"main.?"},
			want:      []string{"main.a", "main.b"},
		},
		{
			name:      "bracket pattern",
			available: []string{"main.a", "main.b", "main.c"},
			patterns:  []string{"main.[ab]"},
			want:      []string{"main.a", "main.b"},
		},
		{
			name:      "deduplicate matches",
			available: []string{"main.gke"},
			patterns:  []string{"main.*", "*.gke"},
			want:      []string{"main.gke"},
		},
		{
			name:      "empty available",
			available: []string{},
			patterns:  []string{"main.*"},
			want:      nil,
		},
		{
			name:      "empty patterns",
			available: []string{"main", "test"},
			patterns:  []string{},
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterNamespaces(tt.available, tt.patterns)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
