package document

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_generateDocument(t *testing.T) {
	tests := []struct {
		name     string
		testdata string
		wantOut  string
		wantErr  bool
	}{
		{
			name:     "Nested packages",
			testdata: "./testdata/foo",
			wantOut:  "./testdata/doc/foo.md",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				as, err := ParseRegoWithAnnotations(tt.testdata)
				assert.NoError(t, err)

				s, err := ConvertAnnotationsToSections(as)
				assert.NoError(t, err)

				gotOut := &bytes.Buffer{}
				err = RenderDocument(gotOut, s)
				if (err != nil) != tt.wantErr {
					t.Errorf("GenVariableDoc() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				wantOut, err := os.ReadFile(tt.wantOut)
				require.NoError(t, err)
				assert.Equal(t, string(wantOut), gotOut.String())

				// un comment this to generate the golden file when changing the template
				os.WriteFile(tt.wantOut+".golden", gotOut.Bytes(), 0644)
			},
		)
	}
}
