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
		name        string
		input       CheckResults
		hideMessage bool
		expected    []string
	}{
		{
			name: "No warnings or failures",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
				},
			},
			expected: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites></testsuites>`,
				``,
			},
		},
		{
			name: "A warning, a failure and a skipped test",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Warnings:  []Result{{Message: "first warning"}},
					Failures:  []Result{{Message: "first failure"}},
					Skipped:   []Result{{Message: "first skipped"}},
				},
			},
			expected: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites>`,
				`	<testsuite tests="3" failures="2" time="0.000" name="conftest.namespace">`,
				`		<properties>`,
				`			<property name="go.version" value="%s"></property>`,
				`		</properties>`,
				`		<testcase classname="conftest.namespace" name="examples/kubernetes/service.yaml - first warning" time="0.000">`,
				`			<failure message="Failed" type="">first warning</failure>`,
				`		</testcase>`,
				`		<testcase classname="conftest.namespace" name="examples/kubernetes/service.yaml - first failure" time="0.000">`,
				`			<failure message="Failed" type="">first failure</failure>`,
				`		</testcase>`,
				`		<testcase classname="conftest.namespace" name="examples/kubernetes/service.yaml - first skipped" time="0.000">`,
				`			<skipped message="first skipped"></skipped>`,
				`		</testcase>`,
				`	</testsuite>`,
				`</testsuites>`,
				``,
			},
		},
		{
			name: "Failure with a long description",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Failures: []Result{{Message: `failure with long message

This is the rest of the description of the failed test`}},
				},
			},
			expected: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites>`,
				`	<testsuite tests="1" failures="1" time="0.000" name="conftest.namespace">`,
				`		<properties>`,
				`			<property name="go.version" value="%s"></property>`,
				`		</properties>`,
				`		<testcase classname="conftest.namespace" name="examples/kubernetes/service.yaml - failure with long message" time="0.000">`,
				`			<failure message="Failed" type="">failure with long message&#xA;&#xA;This is the rest of the description of the failed test</failure>`,
				`		</testcase>`,
				`	</testsuite>`,
				`</testsuites>`,
				``,
			},
		},
		{
			name:        "Failure with --junit-hide-message set",
			hideMessage: true,
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Failures: []Result{{Message: `failure with long message

This is the rest of the description of the failed test`}},
				},
			},
			expected: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<testsuites>`,
				`	<testsuite tests="1" failures="1" time="0.000" name="conftest.namespace">`,
				`		<properties>`,
				`			<property name="go.version" value="%s"></property>`,
				`		</properties>`,
				`		<testcase classname="conftest.namespace" name="examples/kubernetes/service.yaml" time="0.000">`,
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
			expected := strings.Join(tt.expected, "\n")
			if strings.Contains(expected, `%s`) {
				expected = fmt.Sprintf(expected, runtime.Version())
			}

			buf := new(bytes.Buffer)
			if err := NewJUnit(buf, tt.hideMessage).Output(tt.input); err != nil {
				t.Fatal("output junit:", err)
			}
			actual := buf.String()

			if expected != actual {
				t.Errorf("Unexpected output. have:\n %v want:\n %v", actual, expected)
			}
		})
	}
}
