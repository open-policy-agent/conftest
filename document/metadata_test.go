package document

import (
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func validateAnnotation(t *testing.T, c *ast.Compiler, want []string) {
	t.Helper()

	var got []string

	annotations := c.GetAnnotationSet().Flatten()

	for _, entry := range annotations {
		got = append(got, fmt.Sprintf("%s/%s", entry.Annotations.Scope, entry.Annotations.Title))
	}

	assert.ElementsMatch(t, want, got)
}

func TestGetAnnotations(t *testing.T) {
	type args struct {
		directory string
	}
	tests := []struct {
		name string
		args args
		// list of scope/title of the annotation you expect to see
		want    []string
		wantErr bool
	}{
		{
			name: "parse rule level metadata",
			args: args{
				directory: "testdata/foo",
			},
			want: []string{
				"subpackages/foo",
				"package/bar",
				"rule/My Rule A",
				"rule/My Rule P",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRegoWithAnnotations(tt.args.directory)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			validateAnnotation(t, got, tt.want)
		})
	}
}
