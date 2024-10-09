package document

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_generateDocument(t *testing.T) {
	tests := []struct {
		name     string
		testdata string
		Option   []RenderDocumentOption
		wantOut  string
		wantErr  bool
	}{
		{
			name:     "Nested packages",
			testdata: "./testdata/foo",
			wantOut:  "./testdata/doc/foo.md",
			wantErr:  false,
		}, {
			name:     "Nested packages",
			testdata: "./testdata/foo",
			wantOut:  "./testdata/doc/foo.md",
			Option: []RenderDocumentOption{
				WithTemplate("testdata/template.md"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				as, err := ParseRegoWithAnnotations(tt.testdata)
				assert.NoError(t, err)

				d, err := ConvertAnnotationsToSections(as)
				assert.NoError(t, err)

				gotOut := &bytes.Buffer{}
				err = RenderDocument(gotOut, d, tt.Option...)
				if (err != nil) != tt.wantErr {
					t.Errorf("GenVariableDoc() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				wantOut, err := os.ReadFile(tt.wantOut)
				require.NoError(t, err)
				assert.Equal(t, string(wantOut), gotOut.String())

				// prospective golden file, much simpler to see what's the result in case the test fails
				// this does not override the existing test, but create a new file called xxx.golden
				err = os.WriteFile(tt.wantOut+".golden", gotOut.Bytes(), 0600)
				assert.NoError(t, err)
			},
		)
	}
}
