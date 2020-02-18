package commands

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func Test_stdOutputManager_put(t *testing.T) {
	type args struct {
		cr CheckResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    []string
		expErr error
	}{
		{
			msg: "records failure and Warnings",
			args: args{
				cr: CheckResult{
					FileName: "foo.yaml",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult("first failure", []error{})},
				},
			},
			exp: []string{"WARN - foo.yaml - first warning", "FAIL - foo.yaml - first failure"},
		},
		{
			msg: "skips filenames for stdin",
			args: args{
				cr: CheckResult{
					FileName: "-",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult("first failure", []error{})},
				},
			},
			exp: []string{"WARN - first warning", "FAIL - first failure"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewStdOutputManager(log.New(buf, "", 0), false)

			if err := s.Put(tt.args.cr); err != nil {
				t.Fatalf("put output: %v", err)
			}

			actual := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			if !reflect.DeepEqual(tt.exp, actual) {
				t.Errorf("unexpected output. expected %v actual %v", tt.exp, actual)
			}
		})
	}
}

func Test_jsonOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		crs      []CheckResult
	}

	tests := []struct {
		msg  string
		args args
		exp  string
	}{
		{
			msg: "no Warnings or errors",
			args: args{
				crs: []CheckResult{{FileName: "examples/kubernetes/service.yaml"}},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": [],
		"failures": [],
		"successes": []
	}
]
`,
		},
		{
			msg: "records failure and Warnings",
			args: args{
				crs: []CheckResult{{
					FileName: "examples/kubernetes/service.yaml",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult("first failure", []error{})},
				}},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": [
			{
				"info": {
					"msg": "first warning"
				}
			}
		],
		"failures": [
			{
				"info": {
					"msg": "first failure"
				}
			}
		],
		"successes": []
	}
]
`,
		},
		{
			msg: "mixed failure and Warnings",
			args: args{
				crs: []CheckResult{{
					FileName: "examples/kubernetes/service.yaml",
					Failures: []Result{NewResult("first failure", []error{})},
				}},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": [],
		"failures": [
			{
				"info": {
					"msg": "first failure"
				}
			}
		],
		"successes": []
	}
]
`,
		},
		{
			msg: "handles stdin input",
			args: args{
				fileName: "-",
				crs: []CheckResult{{
					Failures: []Result{NewResult("first failure", []error{})}},
				},
			},
			exp: `[
	{
		"filename": "",
		"warnings": [],
		"failures": [
			{
				"info": {
					"msg": "first failure"
				}
			}
		],
		"successes": []
	}
]
`,
		},
		{
			msg: "multiple check results",
			args: args{
				crs: []CheckResult{
					{FileName: "examples/kubernetes/service.yaml"},
					{FileName: "examples/kubernetes/deployment.yaml"},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": [],
		"failures": [],
		"successes": []
	},
	{
		"filename": "examples/kubernetes/deployment.yaml",
		"warnings": [],
		"failures": [],
		"successes": []
	}
]
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewJSONOutputManager(log.New(buf, "", 0))

			for _, cr := range tt.args.crs {
				if err := s.Put(cr); err != nil {
					t.Fatalf("put output: %v", err)
				}
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			if tt.exp != actual {
				t.Errorf("unexpected output. expected %v got %v", tt.exp, actual)
			}
		})
	}
}

func TestSupportedOutputManagers(t *testing.T) {
	for _, testunit := range []struct {
		name          string
		outputFormat  string
		outputManager OutputManager
	}{
		{
			name:          "std output should exist",
			outputFormat:  outputSTD,
			outputManager: NewDefaultStdOutputManager(true),
		},
		{
			name:          "json output should exist",
			outputFormat:  outputJSON,
			outputManager: NewDefaultJSONOutputManager(),
		},
		{
			name:          "tap output should exist",
			outputFormat:  outputTAP,
			outputManager: NewDefaultTAPOutputManager(),
		},
		{
			name:          "table output should exist",
			outputFormat:  outputTable,
			outputManager: NewDefaultTableOutputManager(),
		},
		{
			name:          "default output should exist",
			outputFormat:  "somedefault",
			outputManager: NewDefaultStdOutputManager(true),
		},
	} {
		viper.Set("output", testunit.outputFormat)
		outputManager := GetOutputManager()
		if !reflect.DeepEqual(outputManager, testunit.outputManager) {
			t.Errorf(
				"We expected the output manager to be of type %v : %T and it was %T",
				testunit.outputFormat,
				testunit.outputManager,
				outputManager,
			)
		}

	}
}

func Test_tapOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		cr       CheckResult
	}

	tests := []struct {
		msg  string
		args args
		exp  string
	}{
		{
			msg: "no warnings or errors",
			args: args{
				cr: CheckResult{
					FileName: "examples/kubernetes/service.yaml",
				},
			},
			exp: "",
		},
		{
			msg: "records failure and warnings",
			args: args{
				cr: CheckResult{
					FileName: "examples/kubernetes/service.yaml",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult("first failure", []error{})},
				},
			},
			exp: `1..2
not ok 1 - examples/kubernetes/service.yaml - first failure
# Warnings
not ok 2 - examples/kubernetes/service.yaml - first warning
`,
		},
		{
			msg: "mixed failure and warnings",
			args: args{
				cr: CheckResult{
					FileName: "examples/kubernetes/service.yaml",
					Failures: []Result{NewResult("first failure", []error{})},
				},
			},
			exp: `1..1
not ok 1 - examples/kubernetes/service.yaml - first failure
`,
		},
		{
			msg: "handles stdin input",
			args: args{
				cr: CheckResult{
					FileName: "-",
					Failures: []Result{NewResult("first failure", []error{})},
				},
			},
			exp: `1..1
not ok 1 - first failure
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewTAPOutputManager(log.New(buf, "", 0))

			if err := s.Put(tt.args.cr); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			if tt.exp != actual {
				t.Errorf("unexpected output. expected %v actual %v", tt.exp, actual)
			}
		})
	}
}

func Test_tableOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		cr       CheckResult
	}

	tests := []struct {
		msg  string
		args args
		exp  string
	}{
		{
			msg: "no warnings or errors",
			args: args{
				cr: CheckResult{
					FileName: "examples/kubernetes/service.yaml",
				},
			},
			exp: "",
		},
		{
			msg: "records failure and warnings",
			args: args{
				cr: CheckResult{
					FileName: "examples/kubernetes/service.yaml",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult("first failure", []error{})},
				},
			},
			exp: `+---------+----------------------------------+---------------+
| RESULT  |               FILE               |    MESSAGE    |
+---------+----------------------------------+---------------+
| warning | examples/kubernetes/service.yaml | first warning |
| failure | examples/kubernetes/service.yaml | first failure |
+---------+----------------------------------+---------------+
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewTableOutputManager(buf)

			if err := s.Put(tt.args.cr); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			if tt.exp != actual {
				t.Errorf("unexpected output. expected %v actual %v", tt.exp, actual)
			}
		})
	}
}
