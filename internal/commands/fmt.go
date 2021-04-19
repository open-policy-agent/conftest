package commands

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/open-policy-agent/opa/format"
	"github.com/open-policy-agent/opa/loader"
	"github.com/spf13/cobra"
)

// NewFormatCommand creates a format command.
// This command can be used for formatting Rego files.
func NewFormatCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "fmt <path> [path [...]]",
		Short: "Format Rego files",
		RunE: func(cmd *cobra.Command, files []string) error {
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

				contents, err := ioutil.ReadFile(policy.Package.Location.File)
				if err != nil {
					return fmt.Errorf("read policy: %w", err)
				}

				formattedContents, err := format.Source(policy.Package.Location.File, contents)
				if err != nil {
					return fmt.Errorf("format: %w", err)
				}

				if bytes.Equal(contents, formattedContents) {
					continue
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

	return &cmd
}
