package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/conftest/document"
	"github.com/spf13/cobra"
)

func NewDocumentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc <path> [path [...]]",
		Short: "Generate documentation",
		RunE: func(cmd *cobra.Command, dir []string) error {
			if len(dir) < 1 {
				err := cmd.Usage()
				if err != nil {
					return fmt.Errorf("usage: %s", err)
				}
				return fmt.Errorf("missing required arguments")
			}

			for _, path := range dir {
				fileInfo, err := os.Stat(path)
				if err != nil {
					return err
				}

				if !fileInfo.IsDir() {
					return fmt.Errorf("%s is not a directory", path)
				}

				// Handle the output destination
				outDir, err := cmd.Flags().GetString("outDir")
				if err != nil {
					return fmt.Errorf("invalid outDir: %s", err)
				}

				name := filepath.Base(path)
				if name == "." || name == ".." {
					name = "policy"
				}

				outPath := filepath.Join(outDir, name+".md")
				f, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0600)
				if err != nil {
					return fmt.Errorf("opening %s for writing output: %w", outPath, err)
				}
				defer func(file *os.File) {
					if err := file.Close(); err != nil {
						log.Fatalln(err)
					}
				}(f)

				template, err := cmd.Flags().GetString("template")
				if err != nil {
					return fmt.Errorf("invalid template: %s", err)
				}

				err = document.GenerateDocument(path, template, f)
				if err != nil {
					return fmt.Errorf("generating document: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("outDir", "o", ".", "Path to the output documentation file")
	cmd.Flags().StringP("template", "t", "", "Go template for the document generation")

	return cmd
}
