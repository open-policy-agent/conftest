package output

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestStandard(t *testing.T) {
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
			exp: []string{
				"WARN - foo.yaml - first warning",
				"FAIL - foo.yaml - first failure",
				"",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
			},
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
			exp: []string{
				"WARN - first warning",
				"FAIL - first failure",
				"",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewStandardOutputManager(log.New(buf, "", 0), false)

			if err := s.Put(tt.args.cr); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			if !reflect.DeepEqual(tt.exp, actual) {
				t.Errorf("unexpected output. expected %v actual %v", tt.exp, actual)
			}
		})
	}
}
