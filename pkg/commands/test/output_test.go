package test

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stdOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		cr       checkResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    []string
		expErr error
	}{
		{
			msg: "records failure and warnings",
			args: args{
				fileName: "foo.yaml",
				cr: checkResult{
					warnings: []error{errors.New("first warning")},
					failures: []error{errors.New("first failure")},
				},
			},
			exp: []string{"WARN - foo.yaml - first warning", "FAIL - foo.yaml - first failure"},
		},
		{
			msg: "skips filenames for stdin",
			args: args{
				fileName: "-",
				cr: checkResult{
					warnings: []error{errors.New("first warning")},
					failures: []error{errors.New("first failure")},
				},
			},
			exp: []string{"WARN - first warning", "FAIL - first failure"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := newStdOutputManager(log.New(buf, "", 0), false)

			err := s.put(tt.args.fileName, tt.args.cr)
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
		cr       checkResult
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
				cr:       checkResult{},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": null,
		"failures": null
	}
]
`,
		},
		{
			msg: "records failure and warnings",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr: checkResult{
					warnings: []error{errors.New("first warning")},
					failures: []error{errors.New("first failure")},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": [
			"first warning"
		],
		"failures": [
			"first failure"
		]
	}
]
`,
		},
		{
			msg: "mixed failure and warnings",
			args: args{
				fileName: "examples/kubernetes/service.yaml",
				cr: checkResult{
					failures: []error{errors.New("first failure")},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"warnings": null,
		"failures": [
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
				cr: checkResult{
					failures: []error{errors.New("first failure")},
				},
			},
			exp: `[
	{
		"filename": "",
		"warnings": null,
		"failures": [
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
			s := newJSONOutputManager(log.New(buf, "", 0))

			// record results
			err := s.put(tt.args.fileName, tt.args.cr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// flush final buffer
			err = s.flush()
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			assert.Equal(t, tt.exp, buf.String())
		})
	}
}
