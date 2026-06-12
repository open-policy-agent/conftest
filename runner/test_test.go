package runner

import (
	"reflect"
	"testing"
)

func TestHasGlob(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{name: "empty", patterns: nil, want: false},
		{name: "literals only", patterns: []string{"main", "k8s.simple"}, want: false},
		{name: "single star", patterns: []string{"k8s.*"}, want: true},
		{name: "question mark", patterns: []string{"k8s.simpl?"}, want: true},
		{name: "char class", patterns: []string{"k8s.[a-z]"}, want: true},
		{name: "mixed", patterns: []string{"main", "k8s.*"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasGlob(tt.patterns); got != tt.want {
				t.Errorf("hasGlob(%v) = %v, want %v", tt.patterns, got, tt.want)
			}
		})
	}
}

func TestExpandNamespaceGlobs(t *testing.T) {
	available := []string{
		"main",
		"k8s.simple.deployment",
		"k8s.simple.hpa",
		"k8s.simple.pod",
		"k8s.combined.deployment",
		"data.simple",
	}

	tests := []struct {
		name     string
		patterns []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "literal passes through",
			patterns: []string{"main"},
			want:     []string{"main"},
		},
		{
			name:     "single segment wildcard",
			patterns: []string{"k8s.simple.*"},
			want:     []string{"k8s.simple.deployment", "k8s.simple.hpa", "k8s.simple.pod"},
		},
		{
			name:     "wildcard does not cross dot boundaries",
			patterns: []string{"k8s.*"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "char class",
			patterns: []string{"k8s.simple.[hp]*"},
			want:     []string{"k8s.simple.hpa", "k8s.simple.pod"},
		},
		{
			name:     "mix of literal and glob, dedupes",
			patterns: []string{"main", "k8s.simple.*", "k8s.simple.pod"},
			want:     []string{"main", "k8s.simple.deployment", "k8s.simple.hpa", "k8s.simple.pod"},
		},
		{
			name:     "no match errors",
			patterns: []string{"missing.*"},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandNamespaceGlobs(tt.patterns, available)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
