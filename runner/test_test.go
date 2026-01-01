package runner

import (
	"reflect"
	"sort"
	"testing"
)

func TestFilterNamespaces(t *testing.T) {
	tests := []struct {
		name      string
		available []string
		patterns  []string
		want      []string
	}{
		{
			name:      "exact match single namespace",
			available: []string{"main", "group1", "group2"},
			patterns:  []string{"main"},
			want:      []string{"main"},
		},
		{
			name:      "exact match multiple namespaces",
			available: []string{"main", "group1", "group2"},
			patterns:  []string{"group1", "group2"},
			want:      []string{"group1", "group2"},
		},
		{
			name:      "wildcard match with asterisk",
			available: []string{"main", "group1", "group2", "other"},
			patterns:  []string{"group*"},
			want:      []string{"group1", "group2"},
		},
		{
			name:      "wildcard match all",
			available: []string{"main", "group1", "group2"},
			patterns:  []string{"*"},
			want:      []string{"main", "group1", "group2"},
		},
		{
			name:      "wildcard with question mark",
			available: []string{"group1", "group2", "group10"},
			patterns:  []string{"group?"},
			want:      []string{"group1", "group2"},
		},
		{
			name:      "mixed exact and wildcard",
			available: []string{"main", "group1", "group2", "test"},
			patterns:  []string{"main", "group*"},
			want:      []string{"main", "group1", "group2"},
		},
		{
			name:      "no matches",
			available: []string{"main", "group1"},
			patterns:  []string{"other*"},
			want:      nil,
		},
		{
			name:      "empty patterns",
			available: []string{"main", "group1"},
			patterns:  []string{},
			want:      nil,
		},
		{
			name:      "empty available",
			available: []string{},
			patterns:  []string{"main"},
			want:      nil,
		},
		{
			name:      "no duplicates in result",
			available: []string{"group1", "group2"},
			patterns:  []string{"group1", "group*"},
			want:      []string{"group1", "group2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterNamespaces(tt.available, tt.patterns)
			// Sort both slices for comparison since order may vary
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
