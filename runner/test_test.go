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
			patterns: []string{"main", "kubernetes"},
			want:     false,
		},
		{
			name:     "dotted literal is not a wildcard",
			patterns: []string{"main.gke"},
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
			name:     "wildcard in second pattern",
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
				t.Errorf("hasWildcard(%v) = %v, want %v", tt.patterns, got, tt.want)
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
			name:       "suffix wildcard matches dotted children",
			namespaces: []string{"main", "main.gke", "main.aws", "kubernetes"},
			patterns:   []string{"main.*"},
			want:       []string{"main.gke", "main.aws"},
		},
		{
			name:       "prefix wildcard matches dotted children",
			namespaces: []string{"main.gke", "kubernetes.gke", "main.aws"},
			patterns:   []string{"*.gke"},
			want:       []string{"main.gke", "kubernetes.gke"},
		},
		{
			name:       "literal pattern matches exactly",
			namespaces: []string{"main", "main.gke"},
			patterns:   []string{"main"},
			want:       []string{"main"},
		},
		{
			name:       "literal pattern with no match yields nil",
			namespaces: []string{"main", "kubernetes"},
			patterns:   []string{"notpresent"},
			want:       nil,
		},
		{
			name:       "multiple patterns are unioned",
			namespaces: []string{"main", "main.gke", "terraform", "terraform.aws"},
			patterns:   []string{"main.*", "terraform.*"},
			want:       []string{"main.gke", "terraform.aws"},
		},
		{
			name:       "no matches yields nil",
			namespaces: []string{"main", "kubernetes"},
			patterns:   []string{"foo.*"},
			want:       nil,
		},
		{
			name:       "lone star matches everything",
			namespaces: []string{"main", "kubernetes", "commands"},
			patterns:   []string{"*"},
			want:       []string{"main", "kubernetes", "commands"},
		},
		{
			name:       "question mark matches single character",
			namespaces: []string{"main.a", "main.b", "main.ab"},
			patterns:   []string{"main.?"},
			want:       []string{"main.a", "main.b"},
		},
		{
			name:       "bracket class matches enumerated characters",
			namespaces: []string{"main.a", "main.b", "main.c"},
			patterns:   []string{"main.[ab]"},
			want:       []string{"main.a", "main.b"},
		},
		{
			name:       "overlapping patterns do not duplicate a namespace",
			namespaces: []string{"main.gke"},
			patterns:   []string{"main.*", "*.gke"},
			want:       []string{"main.gke"},
		},
		{
			name:       "result preserves available namespace order",
			namespaces: []string{"zeta", "alpha", "mu"},
			patterns:   []string{"*"},
			want:       []string{"zeta", "alpha", "mu"},
		},
		{
			name:       "empty available yields nil",
			namespaces: []string{},
			patterns:   []string{"main.*"},
			want:       nil,
		},
		{
			name:       "empty patterns yields nil",
			namespaces: []string{"main", "kubernetes"},
			patterns:   []string{},
			want:       nil,
		},
		{
			name:       "invalid pattern returns an error",
			namespaces: []string{"main"},
			patterns:   []string{"[invalid"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := filterNamespaces(tt.namespaces, tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Fatalf("filterNamespaces(%v, %v) error = %v, wantErr %v", tt.namespaces, tt.patterns, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNamespaces(%v, %v) = %v, want %v", tt.namespaces, tt.patterns, got, tt.want)
			}
		})
	}
}
