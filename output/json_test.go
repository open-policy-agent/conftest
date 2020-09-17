package output

import (
	"bytes"
	"log"
	"testing"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []CheckResult
		expected string
	}{
		{
			name: "no warnings or errors",
			input: []CheckResult{
				{FileName: "testdata/kubernetes/service.yaml"},
			},
			expected: `[
	{
		"filename": "testdata/kubernetes/service.yaml",
		"successes": 0
	}
]
`,
		},
		{
			name: "records failures and warnings",
			input: []CheckResult{
				{
					FileName: "testdata/kubernetes/service.yaml",
					Warnings: []Result{{Message: "first warning"}},
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: `[
	{
		"filename": "testdata/kubernetes/service.yaml",
		"successes": 0,
		"warnings": [
			{
				"msg": "first warning"
			}
		],
		"failures": [
			{
				"msg": "first failure"
			}
		]
	}
]
`,
		},
		{
			name: "mixed failure and Warnings",
			input: []CheckResult{
				{
					FileName: "testdata/kubernetes/service.yaml",
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: `[
	{
		"filename": "testdata/kubernetes/service.yaml",
		"successes": 0,
		"failures": [
			{
				"msg": "first failure"
			}
		]
	}
]
`,
		},
		{
			name: "handles stdin input",
			input: []CheckResult{
				{
					FileName: "-",
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: `[
	{
		"filename": "",
		"successes": 0,
		"failures": [
			{
				"msg": "first failure"
			}
		]
	}
]
`,
		},
		{
			name: "multiple check results",
			input: []CheckResult{
				{FileName: "testdata/kubernetes/service.yaml"},
				{FileName: "testdata/kubernetes/deployment.yaml"},
			},
			expected: `[
	{
		"filename": "testdata/kubernetes/service.yaml",
		"successes": 0
	},
	{
		"filename": "testdata/kubernetes/deployment.yaml",
		"successes": 0
	}
]
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewJSONOutputManager(log.New(buf, "", 0))

			for _, cr := range tt.input {
				if err := s.Put(cr); err != nil {
					t.Fatalf("put output: %v", err)
				}
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()
			if tt.expected != actual {
				t.Errorf("unexpected output. expected %v got %v", tt.expected, actual)
			}
		})
	}
}
