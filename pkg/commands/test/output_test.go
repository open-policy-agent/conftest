package test_test

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/instrumenta/conftest/pkg/commands/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_stdOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		cr       test.CheckResult
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
				fileName: "foo.yaml",
				cr: test.CheckResult{
					Warnings: []error{errors.New("first warning")},
					Failures: []error{errors.New("first failure")},
				},
			},
			exp: []string{"WARN - foo.yaml - first warning", "FAIL - foo.yaml - first failure"},
		},
		{
			msg: "skips filenames for stdin",
			args: args{
				fileName: "-",
				cr: test.CheckResult{
					Warnings: []error{errors.New("first warning")},
					Failures: []error{errors.New("first failure")},
				},
			},
			exp: []string{"WARN - first warning", "FAIL - first failure"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := test.NewStdOutputManager(log.New(buf, "", 0), false)

			err := s.Put(tt.args.fileName, tt.args.cr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// split on newlines but remove last one for easier comparisons
			res := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			assert.Equal(t, tt.exp, res)
		})
	}
}

func Test_jsonOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		cr       test.CheckResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    string
		expErr error
	}{
		{
			msg: "no Warnings or errors",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr:       test.CheckResult{},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"Warnings": [],
		"Failures": [],
		"Successes": []
	}
]
`,
		},
		{
			msg: "records failure and Warnings",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr: test.CheckResult{
					Warnings: []error{errors.New("first warning")},
					Failures: []error{errors.New("first failure")},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"Warnings": [
			"first warning"
		],
		"Failures": [
			"first failure"
		],
		"Successes": []
	}
]
`,
		},
		{
			msg: "mixed failure and Warnings",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr: test.CheckResult{
					Failures: []error{errors.New("first failure")},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"Warnings": [],
		"Failures": [
			"first failure"
		],
		"Successes": []
	}
]
`,
		},
		{
			msg: "handles stdin input",
			args: args{
				fileName: "-",
				cr: test.CheckResult{
					Failures: []error{errors.New("first failure")},
				},
			},
			exp: `[
	{
		"filename": "",
		"Warnings": [],
		"Failures": [
			"first failure"
		],
		"Successes": []
	}
]
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := test.NewJSONOutputManager(log.New(buf, "", 0))

			// record results
			err := s.Put(tt.args.fileName, tt.args.cr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// flush final buffer
			err = s.Flush()
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			assert.Equal(t, tt.exp, buf.String())
		})
	}
}

func TestSupportedOutputManagers(t *testing.T) {
	for _, testunit := range []struct {
		name          string
		outputFormat  string
		outputManager test.OutputManager
	}{
		{
			name:          "std output should exist",
			outputFormat:  test.OutputSTD,
			outputManager: test.NewDefaultStdOutputManager(true),
		},
		{
			name:          "json output should exist",
			outputFormat:  test.OutputJSON,
			outputManager: test.NewDefaultJSONOutputManager(),
		},
		{
			name:          "tap output should exist",
			outputFormat:  test.OutputTAP,
			outputManager: test.NewDefaultTAPOutputManager(),
		},
		{
			name:          "default output should exist",
			outputFormat:  "somedefault",
			outputManager: test.NewDefaultStdOutputManager(true),
		},
	} {
		viper.Set("output", testunit.outputFormat)
		outputManager := test.GetOutputManager()
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
		cr       test.CheckResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    string
		expErr error
	}{
		{
			msg: "no warnings or errors",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr:       test.CheckResult{},
			},
			exp: "",
		},
		{
			msg: "records failure and warnings",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr: test.CheckResult{
					Warnings: []error{errors.New("first warning")},
					Failures: []error{errors.New("first failure")},
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
				fileName: "examples/kubernetes/service.yaml",
				cr: test.CheckResult{
					Failures: []error{errors.New("first failure")},
				},
			},
			exp: `1..1
not ok 1 - examples/kubernetes/service.yaml - first failure
`,
		},
		{
			msg: "handles stdin input",
			args: args{
				fileName: "-",
				cr: test.CheckResult{
					Failures: []error{errors.New("first failure")},
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
			s := test.NewTAPOutputManager(log.New(buf, "", 0))

			// record results
			err := s.Put(tt.args.fileName, tt.args.cr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// flush final buffer
			err = s.Flush()
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}
			assert.Equal(t, tt.exp, buf.String())
		})
	}
}
