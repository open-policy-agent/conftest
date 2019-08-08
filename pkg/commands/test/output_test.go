package test_test

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/instrumenta/conftest/pkg/commands/test"
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
		"Failures": []
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
		]
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
		]
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
		]
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
