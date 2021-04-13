package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// Table represents an Outputter that outputs
// results in a tabular format.
type Table struct {
	Writer io.Writer
}

// NewTable creates a new Table with the given writer.
func NewTable(w io.Writer) *Table {
	table := Table{
		Writer: w,
	}

	return &table
}

// Output outputs the results.
func (t *Table) Output(checkResults []CheckResult) error {
	table := tablewriter.NewWriter(t.Writer)
	table.SetHeader([]string{"result", "file", "namespace", "message"})

	for _, checkResult := range checkResults {
		for r := 0; r < checkResult.Successes; r++ {
			table.Append([]string{"success", checkResult.FileName, checkResult.Namespace, "SUCCESS"})
		}


		for _, result := range checkResult.Exceptions {
			table.Append([]string{"exception", checkResult.FileName, checkResult.Namespace, result.Message})
		}

		for _, result := range checkResult.Warnings {
			table.Append([]string{"warning", checkResult.FileName, checkResult.Namespace, result.Message})
		}

		for _, result := range checkResult.Skipped {
			table.Append([]string{"skipped", checkResult.FileName, checkResult.Namespace, result.Message})
		}

		for _, result := range checkResult.Failures {
			table.Append([]string{"failure", checkResult.FileName, checkResult.Namespace, result.Message})
		}
	}

	if table.NumLines() > 0 {
		table.Render()
	}

	return nil
}
