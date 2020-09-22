package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []CheckResult
		expected []string
	}{
		{
			name: "No warnings or errors",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
				},
			},
			expected: []string{
				`[`,
				`	{`,
				`		"filename": "examples/kubernetes/service.yaml",`,
				`		"successes": 0`,
				`	}`,
				`]`,
				``,
			},
		},
		{
			name: "A single failure",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				`[`,
				`	{`,
				`		"filename": "examples/kubernetes/service.yaml",`,
				`		"successes": 0,`,
				`		"failures": [`,
				`			{`,
				`				"msg": "first failure"`,
				`			}`,
				`		]`,
				`	}`,
				`]`,
				``,
			},
		},
		{
			name: "A warning and a failure",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
					Warnings: []Result{{Message: "first warning"}},
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				`[`,
				`	{`,
				`		"filename": "examples/kubernetes/service.yaml",`,
				`		"successes": 0,`,
				`		"warnings": [`,
				`			{`,
				`				"msg": "first warning"`,
				`			}`,
				`		],`,
				`		"failures": [`,
				`			{`,
				`				"msg": "first failure"`,
				`			}`,
				`		]`,
				`	}`,
				`]`,
				``,
			},
		},
		{
			name: "Renames standard input file name to empty string",
			input: []CheckResult{
				{
					FileName: "-",
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				`[`,
				`	{`,
				`		"filename": "",`,
				`		"successes": 0,`,
				`		"failures": [`,
				`			{`,
				`				"msg": "first failure"`,
				`			}`,
				`		]`,
				`	}`,
				`]`,
				``,
			},
		},
		{
			name: "Multiple files",
			input: []CheckResult{
				{FileName: "examples/kubernetes/service.yaml"},
				{FileName: "examples/kubernetes/deployment.yaml"},
			},
			expected: []string{
				`[`,
				`	{`,
				`		"filename": "examples/kubernetes/service.yaml",`,
				`		"successes": 0`,
				`	},`,
				`	{`,
				`		"filename": "examples/kubernetes/deployment.yaml",`,
				`		"successes": 0`,
				`	}`,
				`]`,
				``,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Join(tt.expected, "\n")

			buf := new(bytes.Buffer)
			if err := NewJSON(buf).Output(tt.input); err != nil {
				t.Fatal("output json:", err)
			}
			actual := buf.String()

			if expected != actual {
				t.Errorf("Unexpected output.expected %v actual %v", expected, actual)
			}
		})
	}
}
