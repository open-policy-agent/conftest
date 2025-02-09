package document

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/open-policy-agent/opa/v1/ast"
)

func getTestModules(t *testing.T, modules [][]string) ast.FlatAnnotationsRefSet {
	t.Helper()

	parsed := make([]*ast.Module, 0, len(modules))
	for _, entry := range modules {
		pm, err := ast.ParseModuleWithOpts(entry[0], entry[1], ast.ParserOptions{ProcessAnnotation: true})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		parsed = append(parsed, pm)
	}

	as, err := ast.BuildAnnotationSet(parsed)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	return as.Flatten()
}

// PartialEqual asserts that two objects are equal, depending on what equal means
// For instance, you may pass options to ignore certain fields
// Also, if a struct exports an Equal func this will be used for the assertion
func PartialEqual(t *testing.T, expected, actual any, diffOpts cmp.Option) {
	t.Helper()

	if cmp.Equal(expected, actual, diffOpts) {
		return
	}

	diff := cmp.Diff(expected, actual, diffOpts)

	t.Errorf("Not equal: \n"+
		"expected: %s\n"+
		"actual  : %s%s", expected, actual, diff)
}

func TestParseRegoWithAnnotations(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		// list of scope/title of the annotation you expect to see
		want    []string
		wantErr error
	}{
		{
			name:      "parse package and sub package",
			directory: "testdata/foo",
			want: []string{
				"data.foo",
				"data.foo.a",
				"data.foo.bar",
				"data.foo.bar.p",
			},
		}, {
			name:      "target subpackage",
			directory: "testdata/foo/bar",
			want: []string{
				"data.foo.bar",
				"data.foo.bar.p",
			},
		}, {
			name:      "target example awssam that as no annotations",
			directory: "../examples/awssam",
			want:      []string{},
			wantErr:   ErrNoAnnotations,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ParseRegoWithAnnotations(tt.directory)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("GetAnnotations() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			for i, want := range tt.want {
				if got[i].Path.String() != want {
					t.Errorf("got[%d]Path.String() = %v, want %v", i, tt.want[i], got[i].Path.String())
				}
			}
		})
	}
}

func TestConvertAnnotationsToSection(t *testing.T) {
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
					MarkdownHeading: "#",
					RegoPackageName: "foo",
					Annotations: &ast.Annotations{
						Title: "My Package foo",
					},
				},
				{
					MarkdownHeading: "##",
					RegoPackageName: "foo.p",
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
					MarkdownHeading: "#",
					RegoPackageName: "foo.bar",
					Annotations: &ast.Annotations{
						Title: "My Package bar",
					},
				},
				{
					MarkdownHeading: "##",
					RegoPackageName: "foo.bar.p",
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
					MarkdownHeading: "#",
					RegoPackageName: "foo",
					Annotations: &ast.Annotations{
						Title: "My Package foo",
					},
				},
				{
					MarkdownHeading: "##",
					RegoPackageName: "foo.p",
					Annotations: &ast.Annotations{
						Title: "My Rule P",
					},
				},
				{
					MarkdownHeading: "##",
					RegoPackageName: "foo.q",
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
					MarkdownHeading: "#",
					RegoPackageName: "foo",
					Annotations: &ast.Annotations{
						Title: "My Package foo",
					},
				},
				{
					MarkdownHeading: "##",
					RegoPackageName: "foo.bar",
					Annotations: &ast.Annotations{
						Title: "My Package bar",
					},
				}, {
					MarkdownHeading: "###",
					RegoPackageName: "foo.bar.r",
					Annotations: &ast.Annotations{
						Title: "My Rule R",
					},
				}, {
					MarkdownHeading: "##",
					RegoPackageName: "foo.p",
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
