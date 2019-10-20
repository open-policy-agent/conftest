package parse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/instrumenta/conftest/commands/test"
	"github.com/instrumenta/conftest/parser"
	"github.com/containerd/containerd/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//NewParseCommand creates a parse command.
//This command can be used for printing structured inputs from unstructured configuration inputs.
//Can be used with '--input' or '-i' flag.
func NewParseCommand(osExit func(int)) *cobra.Command {
	ctx := context.Background()
	cmd := &cobra.Command{
		Use:     "parse <filename>",
		Short:   "Print out structured data from your input file",
		PreRun: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlag("input", cmd.Flags().Lookup("input"))
			if err != nil {
				log.G(ctx).Fatal("Failed to bind argument:", err)
			}
		},
		Run: func(cmd *cobra.Command, fileList []string) {
			configurations, err := test.GetConfigurations(ctx, fileList)
			if err != nil {
				log.G(ctx).Fatalf("Your inputs could not be parsed : %s", err)
			}

			for filename, config := range configurations {
				var prettyJSON bytes.Buffer
				os.Stdout.WriteString((filename) + "\n")
				out, err := json.Marshal(config)
				if err != nil {
					log.G(ctx).Fatalf("Error converting config to JSON out: %s", err)
				}
				err = json.Indent(&prettyJSON, out, "", "\t")
				if err != nil {
					log.G(ctx).Fatalf("Error indenting JSON output: %s", err)
				}
				os.Stdout.Write(prettyJSON.Bytes())
			}

			osExit(0)
		},
	}

	cmd.Flags().StringP("input", "i", "", fmt.Sprintf("input type for given source, especially useful when using conftest with stdin, valid options are: %s", parser.ValidInputs()))
	return cmd
}
