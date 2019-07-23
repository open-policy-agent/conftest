package test

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func Test_Multifile(t *testing.T) {
	t.Run("ProcessFile on mulitple files passed", func(t *testing.T) {

		t.Run("it should pass the --combine-yaml-docs flag as an arg", func(t *testing.T) {
			cmd := NewTestCommand()
			if cmd.Flags().Lookup("combine-yaml-docs") == nil {
				t.Errorf("Did not find `--combine-yaml-docs` in command flags. Flags looks like: %v", cmd.Flags())
			}
		})
		t.Run("function should fail if no args", func(t *testing.T) {
			cmd, _, _ := initBasic("../../../testdata/policy")
			args := []string{}

			if os.Getenv("BE_CRASHER") == "1" {
				TestFunction(cmd, args)
				return
			}

			sub := exec.Command(os.Args[0], "-test.run=Test_Multifile")
			sub.Env = append(os.Environ(), "BE_CRASHER=1")
			err := sub.Run()
			if e, ok := err.(*exec.ExitError); !ok || e.Success() {
				t.Fatalf("process ran with err %v, want exit status 1", err)
			}

		})
		t.Run("function should run if there is a arg", func(t *testing.T) {
			cmd, args, _ := initBasic("../../../testdata/policy")
			TestFunction(cmd, args)
		})
		t.Run("function run should if there are multiple args", func(t *testing.T) {
			cmd, args, _ := initBasic("../../../testdata/policy")
			args = append(args, "../../../testdata/weather.yaml")

			TestFunction(cmd, args)
		})

		t.Run("given a concatted yaml file, do we properly handle multiple '---' in the file?", func(t *testing.T) {
			viper.Set("combine-yaml-docs", true)
			yamlFilePath := "../../../testdata/weather-multi-doc.yaml"
			policy := "../../../testdata/policy_with_rules"
			extraArgs := []string{yamlFilePath}
			result := executeTestFunction(policy, extraArgs)

			if strings.TrimSpace(result) != yamlFilePath {
				t.Errorf("Expecting Rego to output; instead got `%s`", result)
			}
		})
	})
}

func initBasic(policy string) (*cobra.Command, []string, context.Context) {
	cmd := &cobra.Command{}
	viper.Set("policy", policy)
	viper.Set("no-color", true)
	viper.Set("namespace", "main")

	ctx := context.Background()

	args := []string{"../../../testdata/name.yaml"}

	return cmd, args, ctx
}

func executeTestFunction(policy string, extraArgs []string) string {
	cmd, _, _ := initBasic(policy)

	//	args = append(args, extraArgs...)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	TestFunction(cmd, extraArgs)

	w.Close()
	os.Stdout = old
	result := <-outC
	return result
}
