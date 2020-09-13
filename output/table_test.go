package output

import (
	"bytes"
	"testing"
)

func TestTable(t *testing.T) {
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
