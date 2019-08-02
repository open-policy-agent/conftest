package report

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"gotest.tools/assert"
)

func TestJSONReporter_Report(t *testing.T) {
	tests := []struct {
		name string
		args chan Result
		exp  []Result
	}{
		{
			"Logs warning messages in correct format",
			getChannel([]Result{
				Result{
					Warn,
					"test",
					"testing warning messages",
				},
			}), 
			[]Result{
				Result{
					Warn,
					"test",
					"testing warning messages",
				},
			},
		},
		{
			"Logs error messages in correct format",
			getChannel([]Result{
				Result{
					Error,
					"test",
					"testing error messages",
				},
			}),
			[]Result{
				Result{
					Error,
					"test",
					"testing error messages",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			reporter := NewJSONReporter(log.New(buf, "", 0))

			err := reporter.Report(tt.args)
			if err != nil {
				t.Errorf("Unexpected error for reporter: %s", err)
			}

			var results []Result
			json.Unmarshal(buf.Bytes(), &results)
			if err != nil {
				t.Errorf("Unexpected error unmarshaling json output: %s", err)
			}

			assert.DeepEqual(t, results, tt.exp)
		})
	}
}
