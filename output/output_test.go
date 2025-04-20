package output

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGetOutputter(t *testing.T) {
	testCases := []struct {
		input    string
		expected Outputter
		tracing  bool
	}{
		{
			input:    OutputStandard,
			expected: NewStandard(os.Stdout),
			tracing:  false,
		},
		{
			input:    OutputStandard,
			expected: NewStandard(os.Stdout),
			tracing:  true,
		},
		{
			input:    OutputJSON,
			expected: NewJSON(os.Stdout),
			tracing:  false,
		},
		{
			input:    OutputJSON,
			expected: NewJSON(os.Stdout),
			tracing:  true,
		},
		{
			input:    OutputTAP,
			expected: NewTAP(os.Stdout),
			tracing:  false,
		},
		{
			input:    OutputTable,
			expected: NewTable(os.Stdout),
			tracing:  false,
		},
		{
			input:    OutputJUnit,
			expected: NewJUnit(os.Stdout, false),
			tracing:  false,
		},
		{
			input:    OutputGitHub,
			expected: NewGitHub(os.Stdout),
			tracing:  false,
		},
		{
			input:    OutputAzureDevOps,
			expected: NewAzureDevOps(os.Stdout),
			tracing:  false,
		},
		{
			input:    OutputSARIF,
			expected: NewSARIF(os.Stdout),
			tracing:  false,
		},
		{
			input:    "unknown_format",
			expected: NewStandard(os.Stdout),
			tracing:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			actual := Get(testCase.input, Options{NoColor: true, Tracing: testCase.tracing})

			// If tracing is enabled, we expect a traceOutputter
			if testCase.tracing {
				if _, ok := actual.(*traceOutputter); !ok {
					t.Errorf("Expected traceOutputter but got %T", actual)
				}
				return
			}

			actualType := reflect.TypeOf(actual)
			expectedType := reflect.TypeOf(testCase.expected)
			if expectedType != actualType {
				t.Errorf("Unexpected outputter. expected %v actual %v", expectedType, actualType)
			}
		})
	}
}

func TestTraceOutputter(t *testing.T) {
	// Create a test result with trace information
	results := CheckResults{
		{
			FileName:  "test.yaml",
			Namespace: "test",
			Failures:  []Result{{Message: "test failure"}},
			Queries: []QueryResult{
				{
					Query: "data.main.deny",
					Traces: []string{
						"TRACE line 1",
						"TRACE line 2",
					},
				},
			},
		},
	}

	tests := []struct {
		name            string
		createOutputter func(io.Writer) Outputter
		expectedStdout  string
		expectedTrace   []string
	}{
		{
			name: "table format with trace",
			createOutputter: func(w io.Writer) Outputter {
				return NewTable(w)
			},
			expectedStdout: "test.yaml",
			expectedTrace:  []string{"TRACE line 1", "TRACE line 2"},
		},
		{
			name: "json format with trace",
			createOutputter: func(w io.Writer) Outputter {
				return NewJSON(w)
			},
			expectedStdout: "test.yaml",
			expectedTrace:  []string{"TRACE line 1", "TRACE line 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffers for stdout and stderr
			stdoutBuf := new(bytes.Buffer)
			stderrBuf := new(bytes.Buffer)

			// Create the main outputter
			mainOutputter := tt.createOutputter(stdoutBuf)

			// Create the trace handler
			traceHandler := &Standard{
				Writer:  stderrBuf,
				NoColor: true,
				Tracing: true,
			}

			// Create the trace outputter
			traceOut := newTraceOutputter(traceHandler, mainOutputter)

			// Output the results
			if err := traceOut.Output(results); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check that trace output went to stderr
			stderrOutput := stderrBuf.String()
			for _, trace := range tt.expectedTrace {
				if !strings.Contains(stderrOutput, trace) {
					t.Errorf("stderr missing expected trace: %q", trace)
				}
			}

			// Check that formatted output went to stdout
			stdoutOutput := stdoutBuf.String()
			if !strings.Contains(stdoutOutput, tt.expectedStdout) {
				t.Errorf("stdout missing expected content: %q", tt.expectedStdout)
			}
		})
	}
}
