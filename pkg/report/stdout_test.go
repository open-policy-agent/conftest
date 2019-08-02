package report

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"gotest.tools/assert"
)

// func TestStdOutReporter_Report(t *testing.T) {

// 	type fields struct {
// 		logger *log.Logger
// 		color  aurora.Aurora
// 	}
// 	type args struct {
// 		level    Level
// 		fileName string
// 		msg      string
// 	}
	// tests := []struct {
	// 	name string
	// 	args args
	// 	exp  string
	// }{
	// 	{
	// 		"Logs warning messages in correct format",
	// 		args{
	// 			Warn,
	// 			"test",
	// 			"testing warning messages",
	// 		},
	// 		"WARN - test - testing warning messages",
	// 	},
	// 	{
	// 		"Logs error messages in correct format",
	// 		args{
	// 			Error,
	// 			"test",
	// 			"testing error messages",
	// 		},
	// 		"FAIL - test - testing error messages",
	// 	},
	// }
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			buf := new(bytes.Buffer)
// 			reporter := NewStdOutReporter(log.New(buf, "", 0), false)

// 			reporter.Report(tt.args.level, tt.args.fileName, tt.args.msg)
// 			res := strings.TrimSuffix(buf.String(), "\n")
// 			assert.Equal(t, res, tt.exp)
// 		})
// 	}
// }

func TestStdOutReporter_Report(t *testing.T) {
	tests := []struct {
		name   string
		args   chan Result
		exp  string
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
			"WARN - test - testing warning messages",
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
			"FAIL - test - testing error messages",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			reporter := NewStdOutReporter(log.New(buf, "", 0), false)

			err := reporter.Report(tt.args)
			if err != nil {
				t.Errorf("Unexpected error for reporter: %s", err)
			}

			res := strings.TrimSuffix(buf.String(), "\n")
			assert.Equal(t, res, tt.exp)
		})
	}
}

func getChannel(results []Result) chan Result {
	resultChan := make(chan Result)
	go func() {
		for _, result := range results {
			resultChan <- result
		}
		close(resultChan)
	}()


	return resultChan
}