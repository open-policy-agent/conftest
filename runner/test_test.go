package runner

import (
	"reflect"
	"testing"
)

func TestHasWildcard(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := hasWildcard(tt.patterns); got != tt.want {
				t.Errorf("hasWildcard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterNamespaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		namespaces []string
		patterns   []string
		want       []string
		wantErr    bool
	}{
		{
			name:       "asterisk matches suffix",
			namespaces: []string{"main", "main.gke", "main.aws", "test"},
			patterns:   []string{"main.*"},
			want:       []string{"main.gke", "main.aws"},
		},
		{
			name:       "asterisk matches prefix",
			namespaces: []string{"main.gke", "test.gke", "main.aws"},
			patterns:   []string{"*.gke"},
			want:       []string{"main.gke", "test.gke"},
		},
		{
			name:       "exact match without wildcard",
			namespaces: []string{"main", "main.gke"},
			patterns:   []string{"main"},
			want:       []string{"main"},
		},
		{
			name:       "multiple patterns",
			namespaces: []string{"main", "main.gke", "test", "test.aws"},
			patterns:   []string{"main.*", "test.*"},
			want:       []string{"main.gke", "test.aws"},
		},
		{
			name:       "no matches",
			namespaces: []string{"main", "test"},
			patterns:   []string{"foo.*"},
			want:       nil,
		},
		{
			name:       "match all with star",
			namespaces: []string{"main", "test", "foo"},
			patterns:   []string{"*"},
			want:       []string{"main", "test", "foo"},
		},
		{
			name:       "question mark pattern",
			namespaces: []string{"main.a", "main.b", "main.ab"},
			patterns:   []string{"main.?"},
			want:       []string{"main.a", "main.b"},
		},
		{
			name:       "bracket pattern",
			namespaces: []string{"main.a", "main.b", "main.c"},
			patterns:   []string{"main.[ab]"},
			want:       []string{"main.a", "main.b"},
		},
		{
			name:       "deduplicate matches",
			namespaces: []string{"main.gke"},
			patterns:   []string{"main.*", "*.gke"},
			want:       []string{"main.gke"},
		},
		{
			name:       "empty available",
			namespaces: []string{},
			patterns:   []string{"main.*"},
			want:       nil,
		},
		{
			name:       "empty patterns",
			namespaces: []string{"main", "test"},
			patterns:   []string{},
			want:       nil,
		},
		{
			name:       "bad pattern",
			namespaces: []string{"main"},
			patterns:   []string{"["}, // Invalid glob pattern
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := filterNamespaces(tt.namespaces, tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
