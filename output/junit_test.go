package output

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

func TestJUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    CheckResult
		expected string
	}{
		{
			name: "no warnings or errors",
			input: CheckResult{
				FileName: "testdata/kubernetes/service.yaml",
			},
			expected: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<testsuites>\n\t<testsuite tests=\"0\" failures=\"0\" time=\"0.000\" name=\"conftest\">\n\t\t<properties>\n\t\t\t<property name=\"go.version\" value=\"%s\"></property>\n\t\t</properties>\n\t</testsuite>\n</testsuites>\n",
		},
		{
			name: "records failure and warnings",
			input: CheckResult{
				FileName: "testdata/kubernetes/service.yaml",
				Warnings: []Result{{Message: "first warning"}},
				Failures: []Result{{Message: "first failure"}},
			},
			expected: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<testsuites>\n\t<testsuite tests=\"2\" failures=\"2\" time=\"0.000\" name=\"conftest\">\n\t\t<properties>\n\t\t\t<property name=\"go.version\" value=\"%s\"></property>\n\t\t</properties>\n\t\t<testcase classname=\"conftest\" name=\"testdata/kubernetes/service.yaml - first warning\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">first warning</failure>\n\t\t</testcase>\n\t\t<testcase classname=\"conftest\" name=\"testdata/kubernetes/service.yaml - first failure\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">first failure</failure>\n\t\t</testcase>\n\t</testsuite>\n</testsuites>\n",
		},
		{
			name: "records failure with long description",
			input: CheckResult{
				FileName: "testdata/kubernetes/service.yaml",
				Warnings: []Result{{Message: "first warning"}},
				Failures: []Result{{Message: `failure with long message

This is the rest of the description of the failed test`},
				}},
			expected: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<testsuites>\n\t<testsuite tests=\"2\" failures=\"2\" time=\"0.000\" name=\"conftest\">\n\t\t<properties>\n\t\t\t<property name=\"go.version\" value=\"%s\"></property>\n\t\t</properties>\n\t\t<testcase classname=\"conftest\" name=\"testdata/kubernetes/service.yaml - first warning\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">first warning</failure>\n\t\t</testcase>\n\t\t<testcase classname=\"conftest\" name=\"testdata/kubernetes/service.yaml - failure with long message\" time=\"0.000\">\n\t\t\t<failure message=\"Failed\" type=\"\">failure with long message&#xA;&#xA;This is the rest of the description of the failed test</failure>\n\t\t</testcase>\n\t</testsuite>\n</testsuites>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewJUnitOutputManager(buf)

			if err := s.Put(tt.input); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			exp := fmt.Sprintf(tt.expected, runtime.Version())
			if exp != actual {
				t.Errorf("unexpected output. expected %s actual %s", tt.expected, actual)
			}
		})
	}
}
