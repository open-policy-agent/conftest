package report

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/logrusorgru/aurora"
)

func TestStdOutReporter_Report(t *testing.T) {

	type fields struct {
		logger *log.Logger
		color  aurora.Aurora
	}
	type args struct {
		level    Level
		fileName string
		msg      string
	}
	tests := []struct {
		name string
		args args
		exp  string
	}{
		{
			"Logs warning messages in correct format",
			args{
				Warn,
				"test",
				"testing warning messages",
			},
			"WARN - test - testing warning messages",
		},
		{
			"Logs error messages in correct format",
			args{
				Error,
				"test",
				"testing error messages",
			},
			"FAIL - test - testing error messages",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			reporter := NewStdOutReporter(log.New(buf, "", 0), false)

			reporter.Report(tt.args.level, tt.args.fileName, tt.args.msg)
			res := strings.TrimSuffix(buf.String(), "\n")
			assert.Equal(t, res, tt.exp)
		})
	}
}
