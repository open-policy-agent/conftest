package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/format"
	"github.com/open-policy-agent/opa/loader"
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
			if err := viper.BindPFlag("check", cmd.Flags().Lookup("check")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}

			return nil
		},
		RunE: func(_ *cobra.Command, files []string) error {
			policies, err := loader.AllRegos(files)
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

				formattedContents, err := format.Source(policy.Package.Location.File, contents)
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

	return &cmd
}
