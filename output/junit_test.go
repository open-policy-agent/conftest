package output

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

func TestJUnit(t *testing.T) {
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
					Filename: "examples/kubernetes/service.yaml",
				},
			},
			exp: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<testsuites>\n\t<testsuite tests=\"0\" failures=\"0\" time=\"0.000\" name=\"conftest\">\n\t\t<properties>\n\t\t\t<property name=\"go.version\" value=\"%s\"></property>\n\t\t</properties>\n\t</testsuite>\n</testsuites>\n",
		},
		{
			msg: "records failure and warnings",
			args: args{
				cr: CheckResult{
					Filename: "examples/kubernetes/service.yaml",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult("first failure", []error{
						fmt.Errorf("this is an error"),
					})},
				},
			},
			exp: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<testsuites>\n\t<testsuite tests=\"2\" failures=\"2\" time=\"0.000\" name=\"conftest\">\n\t\t<properties>\n\t\t\t<property name=\"go.version\" value=\"%s\"></property>\n\t\t</properties>\n\t\t<testcase classname=\"conftest\" name=\"examples/kubernetes/service.yaml - first warning\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">first warning</failure>\n\t\t</testcase>\n\t\t<testcase classname=\"conftest\" name=\"examples/kubernetes/service.yaml - first failure\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">first failure&#xA;this is an error</failure>\n\t\t</testcase>\n\t</testsuite>\n</testsuites>\n",
		},
		{
			msg: "records failure with long description",
			args: args{
				cr: CheckResult{
					Filename: "examples/kubernetes/service.yaml",
					Warnings: []Result{NewResult("first warning", []error{})},
					Failures: []Result{NewResult(`failure with long message

This is the rest of the description of the failed test`, []error{})},
				},
			},
			exp: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<testsuites>\n\t<testsuite tests=\"2\" failures=\"2\" time=\"0.000\" name=\"conftest\">\n\t\t<properties>\n\t\t\t<property name=\"go.version\" value=\"%s\"></property>\n\t\t</properties>\n\t\t<testcase classname=\"conftest\" name=\"examples/kubernetes/service.yaml - first warning\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">first warning</failure>\n\t\t</testcase>\n\t\t<testcase classname=\"conftest\" name=\"examples/kubernetes/service.yaml - failure with long message\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">failure with long message&#xA;&#xA;This is the rest of the description of the failed test</failure>\n\t\t</testcase>\n\t</testsuite>\n</testsuites>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewJUnitOutputManager(buf)

			if err := s.Put(tt.args.cr); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			exp := fmt.Sprintf(tt.exp, runtime.Version())
			if exp != actual {
				t.Errorf("unexpected output. expected %q actual %q", tt.exp, actual)
			}
		})
	}
}
