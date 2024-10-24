package document

import (
	"bytes"
	"os"
	"testing"
)

func TestRenderDocument(t *testing.T) {
	tests := []struct {
		name     string
		testdata string
		Options  []RenderDocumentOption
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
			Options: []RenderDocumentOption{
				ExternalTemplate("testdata/template.md"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as, err := ParseRegoWithAnnotations(tt.testdata)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegoWithAnnotations() error = %v, wantErr %v", err, tt.wantErr)
			}

			d, err := ConvertAnnotationsToSections(as)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertAnnotationsToSections() error = %v, wantErr %v", err, tt.wantErr)
			}

			gotOut := &bytes.Buffer{}
			err = RenderDocument(gotOut, d, tt.Options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantOut, err := os.ReadFile(tt.wantOut)
			if err != nil {
				t.Errorf("unexpected test error: %v", err)
				return
			}

			if gotOut.String() != string(wantOut) {
				t.Errorf("ReadFile() = %v, want %v", gotOut.String(), wantOut)
			}

			// prospective golden file, much simpler to see what's the result in case the test fails
			// this does not override the existing test, but create a new file called xxx.golden
			err = os.WriteFile(tt.wantOut+".golden", gotOut.Bytes(), 0600)
			if err != nil {
				t.Errorf("unexpected test error: %v", err)
				return
			}
		},
		)
	}
}
