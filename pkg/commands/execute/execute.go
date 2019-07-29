package execute

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/log"
	"github.com/instrumenta/conftest/pkg/compiler"
	"github.com/open-policy-agent/opa/tester"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewExecuteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute <file> [file...]",
		Short: "Execute Rego unit tests",
	
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
	
			compiler, err := compiler.BuildCompiler(viper.GetString("policy"))
			if err != nil {
				log.G(ctx).Fatalf("Problem building rego compiler: %s", err)
			}
	
			runner := tester.NewRunner().
				SetCompiler(compiler)
	
			ch, err := runner.Run(ctx, compiler.Modules)
			if err != nil {
				log.G(ctx).Fatalf("Problem running rego tests: %s", err)
			}
			
			for result := range ch {
				fmt.Println(result)
			}
			
			os.Exit(0)
		},
	}

	return cmd
}
