package output

import (
	"bytes"
	"log"
	"testing"
)

func TestTAP(t *testing.T) {
	type args struct {
		cr CheckResult
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
