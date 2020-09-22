package output

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestJUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    []CheckResult
		expected []string
	}{
		{
			name: "No warnings or failures",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
				},
			},
			expected: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites>`,
				`	<testsuite tests="0" failures="0" time="0.000" name="conftest">`,
				`		<properties>`,
				`			<property name="go.version" value="%s"></property>`,
				`		</properties>`,
				`	</testsuite>`,
				`</testsuites>`,
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
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites>`,
				`	<testsuite tests="2" failures="2" time="0.000" name="conftest">`,
				`		<properties>`,
				`			<property name="go.version" value="%s"></property>`,
				`		</properties>`,
				`		<testcase classname="conftest" name="examples/kubernetes/service.yaml - first warning" time="0.000">`,
				`			<failure message="Failed" type="">first warning</failure>`,
				`		</testcase>`,
				`		<testcase classname="conftest" name="examples/kubernetes/service.yaml - first failure" time="0.000">`,
				`			<failure message="Failed" type="">first failure</failure>`,
				`		</testcase>`,
				`	</testsuite>`,
				`</testsuites>`,
				``,
			},
		},
		{
			name: "Failure with a long description",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
					Failures: []Result{{Message: `failure with long message

This is the rest of the description of the failed test`}},
				},
			},
			expected: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites>`,
				`	<testsuite tests="1" failures="1" time="0.000" name="conftest">`,
				`		<properties>`,
				`			<property name="go.version" value="%s"></property>`,
				`		</properties>`,
				`		<testcase classname="conftest" name="examples/kubernetes/service.yaml - failure with long message" time="0.000">`,
				`			<failure message="Failed" type="">failure with long message&#xA;&#xA;This is the rest of the description of the failed test</failure>`,
				`		</testcase>`,
				`	</testsuite>`,
				`</testsuites>`,
				``,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := fmt.Sprintf(strings.Join(tt.expected, "\n"), runtime.Version())

			buf := new(bytes.Buffer)
			if err := NewJUnit(buf).Output(tt.input); err != nil {
				t.Fatal("output junit:", err)
			}
			actual := buf.String()

			if expected != actual {
				t.Errorf("Unexpected output. expected %v actual %v", expected, actual)
			}
		})
	}
}
