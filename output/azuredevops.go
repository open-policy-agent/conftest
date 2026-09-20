package output

import (
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/v1/tester"
)

// AzureDevOps represents an Outputter that outputs
// results in AzureDevOps Pipelines format.
// https://learn.microsoft.com/en-us/azure/devops/pipelines/scripts/logging-commands
type AzureDevOps struct {
	writer io.Writer
}

// NewAzureDevOps creates a new AzureDevOps with the given writer.
func NewAzureDevOps(w io.Writer) *AzureDevOps {
	azuredevops := AzureDevOps{
		writer: w,
	}

	return &azuredevops
}

// Output outputs the results.
func (t *AzureDevOps) Output(checkResults CheckResults) error {
	for _, result := range checkResults {
		totalPolicies := result.Totals().Tests()

		fmt.Fprintf(t.writer, "##[section]Testing '%s' against %d policies in namespace '%s'\n", result.FileName, totalPolicies, result.Namespace)
		fmt.Fprintf(t.writer, "##[group]See conftest results\n")
		for _, failure := range result.Failures {
			fmt.Fprintf(t.writer, "##vso[task.logissue type=error] file=%s --> %s\n", result.FileName, failure.Message)
		}

		for _, warning := range result.Warnings {
			fmt.Fprintf(t.writer, "##vso[task.logissue type=warning] file=%s --> %s\n", result.FileName, warning.Message)
		}

		for _, exception := range result.Exceptions {
			fmt.Fprintf(t.writer, "##vso[task.logissue type=warning] file=%s --> %s\n", result.FileName, exception.Message)
		}

		for _, skipped := range result.Skipped {
			fmt.Fprintf(t.writer, "skipped file=%s %s\n", result.FileName, skipped.Message)
		}

		if result.Successes > 0 {
			fmt.Fprintf(t.writer, "success file=%s %d\n", result.FileName, result.Successes)
		}

		fmt.Fprintf(t.writer, "##[endgroup]\n")
	}

	fmt.Fprintln(t.writer, checkResults.Totals())

	return nil
}

func (t *AzureDevOps) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in AzureDevOps output")
}
