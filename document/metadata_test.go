package document

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateAnnotation(t *testing.T, as ast.FlatAnnotationsRefSet, want []string) {
	t.Helper()

	var got []string

	for _, entry := range as {
		got = append(got, fmt.Sprintf("%s/%s", entry.Annotations.Scope, entry.Annotations.Title))
	}

	assert.ElementsMatch(t, want, got)
}

func getTestModules(t *testing.T, modules [][]string) ast.FlatAnnotationsRefSet {
	t.Helper()

	parsed := make([]*ast.Module, 0, len(modules))
	for _, entry := range modules {
		pm, err := ast.ParseModuleWithOpts(entry[0], entry[1], ast.ParserOptions{ProcessAnnotation: true})
		require.NoError(t, err)
		parsed = append(parsed, pm)
	}

	as, err := ast.BuildAnnotationSet(parsed)
	require.Nil(t, err)

	return as.Flatten()
}

// PartialEqual asserts that two objects are equal, depending on what equal means
// For instance, you may pass options to ignore certain fields
// Also, if a struct exports an Equal func this will be used for the assertion
func PartialEqual(t *testing.T, expected, actual any, diffOpts cmp.Option, msgAndArgs ...any) {
	t.Helper()

	if cmp.Equal(expected, actual, diffOpts) {
		return
	}

	diff := cmp.Diff(expected, actual, diffOpts)
	assert.Fail(t, fmt.Sprintf("Not equal: \n"+
		"expected: %s\n"+
		"actual  : %s%s", expected, actual, diff), msgAndArgs...)
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
				"subpackages/My package foo",
				"package/My package bar",
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

func TestGetDocument(t *testing.T) {

	tests := []struct {
		name    string
		modules [][]string
		want    Document
		wantErr bool
	}{
		{
			name: "Single file",
			modules: [][]string{
				{"foo.rego", `
# METADATA 
# title: My Package foo
package foo

# METADATA 
# title: My Rule P
p := 7
`},
			},
			want: Document{
				{
					H:    "#",
					Path: "foo",
					Annotations: &ast.Annotations{
						Title: "My Package foo",
					},
				},
				{
					H:    "##",
					Path: "foo.p",
					Annotations: &ast.Annotations{
						Title: "My Rule P",
					},
				},
			},
		},
		{
			name: "Single file of a subpackage",
			modules: [][]string{
				{"foo/bar.rego", `
# METADATA 
# title: My Package bar
package foo.bar

# METADATA 
# title: My Rule P
p := 7
`},
			},
			want: Document{
				{
					H:    "#",
					Path: "foo.bar",
					Annotations: &ast.Annotations{
						Title: "My Package bar",
					},
				},
				{
					H:    "##",
					Path: "foo.bar.p",
					Annotations: &ast.Annotations{
						Title: "My Rule P",
					},
				},
			},
		},
		{
			name: "Single file, multiple rule and package metadata",
			modules: [][]string{
				{"foo.rego", `
# METADATA 
# title: My Package foo
package foo

# METADATA 
# title: My Rule P
p := 7

# METADATA 
# title: My Rule Q
q := 8
`},
			},
			want: Document{
				{
					H:    "#",
					Path: "foo",
					Annotations: &ast.Annotations{
						Title: "My Package foo",
					},
				},
				{
					H:    "##",
					Path: "foo.p",
					Annotations: &ast.Annotations{
						Title: "My Rule P",
					},
				},
				{
					H:    "##",
					Path: "foo.q",
					Annotations: &ast.Annotations{
						Title: "My Rule Q",
					},
				},
			},
		}, {
			name: "Multiple file and subpackage",
			modules: [][]string{
				{"foo.rego", `
# METADATA 
# title: My Package foo
package foo

# METADATA 
# title: My Rule P
p := 7

`},
				{"bar/bar.rego", `
# METADATA 
# title: My Package bar
package foo.bar

# METADATA 
# title: My Rule R
r := 9

`},
			},
			want: Document{
				{
					H:    "#",
					Path: "foo",
					Annotations: &ast.Annotations{
						Title: "My Package foo",
					},
				},
				{
					H:    "##",
					Path: "foo.bar",
					Annotations: &ast.Annotations{
						Title: "My Package bar",
					},
				}, {
					H:    "###",
					Path: "foo.bar.r",
					Annotations: &ast.Annotations{
						Title: "My Rule R",
					},
				}, {
					H:    "##",
					Path: "foo.p",
					Annotations: &ast.Annotations{
						Title: "My Rule P",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := getTestModules(t, tt.modules)
			got, err := ConvertAnnotationsToSections(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertAnnotationsToSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			PartialEqual(t, tt.want, got, nil)
		})
	}
}
