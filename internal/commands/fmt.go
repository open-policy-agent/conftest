package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/format"
	"github.com/open-policy-agent/opa/v1/loader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewFormatCommand creates a format command.
// This command can be used for formatting Rego files.
func NewFormatCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "fmt <path> [path [...]]",
		Short: "Format Rego files",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			flagNames := []string{
				"check",
				"rego-version",
			}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},
		RunE: func(_ *cobra.Command, files []string) error {
			selectedRegoVersion, err := parseRegoVersion(viper.GetString("rego-version"))
			if err != nil {
				return fmt.Errorf("failed to parse rego-version flag: %w", err)
			}

			policies, err := loader.NewFileLoader().WithRegoVersion(selectedRegoVersion).Filtered(files, func(_ string, info os.FileInfo, _ int) bool {
				return !info.IsDir() && !strings.HasSuffix(info.Name(), bundle.RegoExt)
			})
			if err != nil {
				return fmt.Errorf("get rego files: %w", err)
			} else if len(policies.Modules) == 0 {
				return fmt.Errorf("no policies found in %v", files)
			}

			for _, policy := range policies.ParsedModules() {
				info, err := os.Stat(policy.Package.Location.File)
				if err != nil {
					return fmt.Errorf("stat: %w", err)
				}

				contents, err := os.ReadFile(policy.Package.Location.File)
				if err != nil {
					return fmt.Errorf("read policy: %w", err)
				}

				formatOpts := format.Opts{RegoVersion: selectedRegoVersion, ParserOptions: &ast.ParserOptions{RegoVersion: selectedRegoVersion}}
				formattedContents, err := format.SourceWithOpts(policy.Package.Location.File, contents, formatOpts)
				if err != nil {
					return fmt.Errorf("format: %w", err)
				}

				// If the original file contents match the formatted contents, formatting does not
				// need to be done and we can try the next module.
				if bytes.Equal(contents, formattedContents) {
					continue
				}

				// When we are running the format command in check mode and the file contents are different
				// we want to return an error code to the user and not update any of the files.
				if viper.GetBool("check") {
					return errors.New("files not formatted")
				}

				outfile, err := os.OpenFile(policy.Package.Location.File, os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
				if err != nil {
					return fmt.Errorf("open file for write: %w", err)
				}

				if _, err = outfile.Write(formattedContents); err != nil {
					return fmt.Errorf("write formatted contents: %w", err)
				}

				outfile.Close()
			}

			return nil
		},
	}

	cmd.Flags().Bool("check", false, "Returns a non-zero exit code if the policies are not formatted")
	cmd.Flags().String("rego-version", "v1", "Which version of Rego syntax to use. Options: v0, v1")

	return &cmd
}

func parseRegoVersion(regoVersionStr string) (ast.RegoVersion, error) {
	switch regoVersionStr {
	case "v0", "V0":
		return ast.RegoV0, nil
	case "v1", "V1":
		return ast.RegoV1, nil
	default:
		return -1, fmt.Errorf("invalid Rego version: %s", regoVersionStr)
	}
}
