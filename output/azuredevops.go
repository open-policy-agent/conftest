package output

import (
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/tester"
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
func (t *AzureDevOps) Output(checkResults []CheckResult) error {
	var totalFailures int
	var totalExceptions int
	var totalWarnings int
	var totalSuccesses int
	var totalSkipped int

	for _, result := range checkResults {
		totalPolicies := result.Successes + len(result.Failures) + len(result.Warnings) + len(result.Exceptions) + len(result.Skipped)

		fmt.Fprintf(t.writer, "##[section]Testing '%v' against %v policies in namespace '%v'\n", result.FileName, totalPolicies, result.Namespace)
		fmt.Fprintf(t.writer, "##[group]See conftest results\n")
		for _, failure := range result.Failures {
			fmt.Fprintf(t.writer, "##vso[task.logissue type=error] file=%v --> %v\n", result.FileName, failure.Message)
		}

		for _, warning := range result.Warnings {
			fmt.Fprintf(t.writer, "##vso[task.logissue type=warning] file=%v --> %v\n", result.FileName, warning.Message)
		}

		for _, exception := range result.Exceptions {
			fmt.Fprintf(t.writer, "##vso[task.logissue type=warning] file=%v --> %v\n", result.FileName, exception.Message)
		}

		for _, skipped := range result.Skipped {
			fmt.Fprintf(t.writer, "skipped file=%v %v\n", result.FileName, skipped.Message)
		}

		if result.Successes > 0 {
			fmt.Fprintf(t.writer, "success file=%v %v\n", result.FileName, result.Successes)
		}

		totalFailures += len(result.Failures)
		totalExceptions += len(result.Exceptions)
		totalWarnings += len(result.Warnings)
		totalSkipped += len(result.Skipped)
		totalSuccesses += result.Successes

		fmt.Fprintf(t.writer, "##[endgroup]\n")
	}

	totalTests := totalFailures + totalExceptions + totalWarnings + totalSuccesses + totalSkipped

	var pluralSuffixTests string
	if totalTests != 1 {
		pluralSuffixTests = "s"
	}

	var pluralSuffixWarnings string
	if totalWarnings != 1 {
		pluralSuffixWarnings = "s"
	}

	var pluralSuffixFailures string
	if totalFailures != 1 {
		pluralSuffixFailures = "s"
	}

	var pluralSuffixExceptions string
	if totalExceptions != 1 {
		pluralSuffixExceptions = "s"
	}

	outputText := fmt.Sprintf("%v test%s, %v passed, %v warning%s, %v failure%s, %v exception%s",
		totalTests, pluralSuffixTests,
		totalSuccesses,
		totalWarnings, pluralSuffixWarnings,
		totalFailures, pluralSuffixFailures,
		totalExceptions, pluralSuffixExceptions,
	)
	fmt.Fprintln(t.writer, outputText)

	return nil
}

func (t *AzureDevOps) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in AzureDevOps output")
}
