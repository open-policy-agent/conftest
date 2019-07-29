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

func TestSingleFileTF(t *testing.T) {
	t.Run("given a single terraform file and a valid policy", func(t *testing.T) {
		fileList := []string{"../../../testdata/tf_single_file.tf"}
		viper.Set("policy", "../../../testdata/policy")
		viper.Set("no-color", true)
		viper.Set("namespace", "main")
		t.Run("when a combine-files flag is false", func(t *testing.T) {
			viper.Set("combine-files", false)
			exitSpyCalled := 0
			cmd := NewTestCommand(func(exitCode int) {
				exitSpyCalled += exitCode
			})
			t.Run("then we should be able to check a file with a single policy", func(t *testing.T) {
				cmd.Run(cmd, fileList)
				if exitSpyCalled != 0 {
					t.Errorf("we failed out of the policy run with exitcode: %v", exitSpyCalled)
				}
			})
		})
	})
}

func TestMultiFile(t *testing.T) {
	t.Run("given multiple files and a policy which is met across files", func(t *testing.T) {

		fileList := []string{
			"../../../testdata/multi_file_part_1.tf",
			"../../../testdata/multi_file_part_2.tf",
		}
		viper.Set("policy", "../../../testdata/policy")
		viper.Set("no-color", true)
		viper.Set("namespace", "main")
		t.Run("when a combine-files flag is true", func(t *testing.T) {
			viper.Set("combine-files", true)
			viper.Set("namespace", "combine")
			exitSpyCalled := 0
			cmd := NewTestCommand(func(exitCode int) {
				exitSpyCalled += exitCode
			})
			t.Run("then we should be able to check across files with a single policy", func(t *testing.T) {
				cmd.Run(cmd, fileList)
				if exitSpyCalled != 0 {
					t.Errorf("we failed out of the policy run with exitcode: %v", exitSpyCalled)
				}
			})
		})

		t.Run("when a combine-files flag is false", func(t *testing.T) {
			viper.Set("combine-files", false)
			exitSpyCalled := 0
			cmd := NewTestCommand(func(exitCode int) {
				exitSpyCalled += exitCode
			})
			t.Run("then we should NOT be able to check across files with a single policy", func(t *testing.T) {
				cmd.Run(cmd, fileList)
				if exitSpyCalled == 0 {
					t.Errorf("we should not have passed here, but we did")
				}
			})
		})
	})
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
