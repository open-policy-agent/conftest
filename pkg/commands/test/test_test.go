package test

import (
	"testing"

	"github.com/spf13/viper"
)

func TestWarnQuery(t *testing.T) {

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

func TestCombineConfig(t *testing.T) {
	viper.Set("namespace", "main")
	t.Run("combine-config flag exists", func(t *testing.T) {
		callCount := 0
		cmd := NewTestCommand(func(int) {
			callCount += 1
		})
		if cmd.Flag("combine-config") == nil {
			t.Errorf("combine-config flag should exist")
		}
	})

	t.Run("given a valid policy and a valid config", func(t *testing.T) {
		viper.Set("combine-config", false)
		viper.Set("policy", "testdata/policy")
		callCount := 0
		cmd := NewTestCommand(func(int) {
			callCount += 1
		})
		cmd.Run(cmd, []string{"testdata/deployment.yaml"})
		if callCount > 0 {
			t.Errorf("we exited with a failure: %v", callCount)
		}
	})

	t.Run("given a valid policy and multiple configs", func(t *testing.T) {
		viper.Set("combine-config", false)
		viper.Set("policy", "testdata/policy")
		callCount := 0
		cmd := NewTestCommand(func(int) {
			callCount += 1
		})
		cmd.Run(cmd, []string{"testdata/deployment+service.yaml", "testdata/deployment.yaml"})
		if callCount > 0 {
			t.Errorf("we exited with a failure: %v", callCount)
		}
	})

	t.Run("given a valid policy multiple configs and `combine-config` flag set to true", func(t *testing.T) {
		t.Skip("not yet implemented")
		viper.Set("combine-config", true)
		viper.Set("policy", "testdata/policy")
		callCount := 0
		cmd := NewTestCommand(func(int) {
			callCount += 1
		})
		cmd.Run(cmd, []string{"testdata/failing_alone.yaml", "testdata/deployment.yaml"})
		if callCount > 0 {
			t.Errorf("we exited with a failure: %v", callCount)
		}
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
