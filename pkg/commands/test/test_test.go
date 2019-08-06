package test

import (
	"testing"

	"github.com/spf13/viper"
)

func TestWarnQuerry(t *testing.T) {

	tests := []struct {
		in  string
		exp bool
	}{
		{"", false},
		{"warn", true},
		{"warnXYZ", false},
		{"warn_", false},
		{"warn_x", true},
		{"warn_x_y_z", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := warnQ.MatchString(tt.in)

			if tt.exp != res {
				t.Fatalf("%s recognized as `warn` query - expected: %v actual: %v", tt.in, tt.exp, res)
			}
		})
	}
}

func TestFailQuery(t *testing.T) {

	tests := []struct {
		in  string
		exp bool
	}{
		{"", false},
		{"deny", true},
		{"denyXYZ", false},
		{"deny_", false},
		{"deny_x", true},
		{"deny_x_y_z", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := denyQ.MatchString(tt.in)

			if tt.exp != res {
				t.Fatalf("%s recognized as `fail` query - expected: %v actual: %v", tt.in, tt.exp, res)
			}
		})
	}
}

func TestMultiFile(t *testing.T) {
	t.Run("Given a rego policy it is evaluated", func(t *testing.T) {
		testTable := []struct {
			name             string
			fileList         []string
			expectedExitCode int
			policyLocation   string
		}{
			{
				name: "Evaluating a rego policy on a single file",
				fileList: []string{
					"testdata/yaml/simple.yaml",
				},
				expectedExitCode: 0,
				policyLocation:   "testData/yaml/simple.rego",
			},
			{
				name: "Evaluating a rego policy against multiple files separately",
				fileList: []string{
					"testdata/yaml/simple.yaml",
					"testdata/yaml/basic.yaml",
				},
				expectedExitCode: 0,
				policyLocation:   "testdata/yaml/simple.rego",
			},
			{
				name: "evaluating a failing rego policy on a single file",
				fileList: []string{
					"testdata/yaml/simpleFail.yaml",
				},
				expectedExitCode: 1,
				policyLocation:   "testData/yaml/simple.rego",
			},
		}

		for _, test := range testTable {
			t.Run(test.name, func(t *testing.T) {
				viper.Set("policy", test.policyLocation)
				viper.Set("namespace", "main")
				exitSpyCalled := 0
				cmd := NewTestCommand(func(exitCode int) {
					exitSpyCalled += exitCode
				})

				cmd.Run(cmd, test.fileList)

				if exitSpyCalled != test.expectedExitCode {
					t.Errorf("We expected exitSpy to be %v but instead we got %v", test.expectedExitCode, exitSpyCalled)
				}
			})
		}

	})
}
