package output

import (
	"fmt"
	"io"
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
func (t *TAP) Output(checkResults []CheckResult) error {
	for _, result := range checkResults {
		var indicator string
		if result.FileName == "-" {
			indicator = "-"
		} else {
			indicator = fmt.Sprintf("- %s -", result.FileName)
		}

		totalTests := result.Successes + len(result.Failures) + len(result.Warnings) + len(result.Exceptions)
		if totalTests == 0 {
			return nil
		}

		counter := 1
		fmt.Fprintln(t.Writer, fmt.Sprintf("1..%d", totalTests))

		for _, failure := range result.Failures {
			fmt.Fprintln(t.Writer, fmt.Sprintf("not ok %v %v %v", counter, indicator, failure.Message))
			counter++
		}

		if len(result.Warnings) > 0 {
			fmt.Fprintln(t.Writer, "# warnings")
			for _, warning := range result.Warnings {
				fmt.Fprintln(t.Writer, fmt.Sprintf("not ok %v %v %v", counter, indicator, warning.Message))
				counter++
			}
		}

		if len(result.Exceptions) > 0 {
			fmt.Fprintln(t.Writer, "# exceptions")
			for _, exception := range result.Exceptions {
				fmt.Fprintln(t.Writer, fmt.Sprintf("ok %v %v %v", counter, indicator, exception.Message))
				counter++
			}
		}

		if result.Successes > 0 {
			fmt.Fprintln(t.Writer, "# successes")
			for i := 0; i < result.Successes; i++ {
				fmt.Fprintln(t.Writer, fmt.Sprintf("ok %v %v", counter, indicator))
				counter++
			}
		}
	}

	return nil
}
