package output

import (
	"fmt"
	"io"
	"slices"

	"github.com/open-policy-agent/opa/v1/tester"
)

// TAP represents an Outputter that outputs
// results in TAP format.
type TAP struct {
	Writer io.Writer
}

// NewTAP creates a new TAP with the given writer.
func NewTAP(w io.Writer) *TAP {
	tap := TAP{
		Writer: w,
	}

	return &tap
}

// Output outputs the results.
func (t *TAP) Output(checkResults CheckResults) error {
	for _, result := range checkResults {
		var indicator string
		var namespace string
		if result.FileName == "-" {
			indicator = "-"
		} else {
			indicator = fmt.Sprintf("- %s", result.FileName)
		}

		if result.Namespace == "-" {
			namespace = "-"
		} else {
			namespace = fmt.Sprintf("- %s -", result.Namespace)
		}

		totalTests := result.Totals().Tests()
		if totalTests == 0 {
			return nil
		}

		fmt.Fprintf(t.Writer, "1..%d\n", totalTests)

		sections := []struct {
			header  string
			status  string
			results []Result
		}{
			{status: "not ok", results: result.Failures},
			{header: "warnings", status: "not ok", results: result.Warnings},
			{header: "exceptions", status: "ok", results: result.Exceptions},
			{header: "skip", status: "ok", results: result.Skipped},
			{header: "successes", status: "ok", results: slices.Repeat([]Result{{Message: "SUCCESS"}}, result.Successes)},
		}

		counter := 1
		for _, section := range sections {
			if section.header != "" && len(section.results) > 0 {
				fmt.Fprintln(t.Writer, "#", section.header)
			}
			for _, r := range section.results {
				fmt.Fprintf(t.Writer, "%s %d %s %s %s\n", section.status, counter, indicator, namespace, r.Message)
				counter++
			}
		}
	}

	return nil
}

func (t *TAP) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in TAP output")
}
