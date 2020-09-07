package output

import (
	"io"
	"os"

	table "github.com/olekukonko/tablewriter"
)

// TableOutputManager formats its output in a table
type TableOutputManager struct {
	table *table.Table
}

// NewDefaultTableOutputManager creates a new TableOutputManager using standard out
func NewDefaultTableOutputManager() *TableOutputManager {
	return NewTableOutputManager(os.Stdout)
}

// NewTableOutputManager creates a new TableOutputManager with a given Writer
func NewTableOutputManager(w io.Writer) *TableOutputManager {
	table := table.NewWriter(w)
	table.SetHeader([]string{"result", "file", "message"})
	return &TableOutputManager{
		table: table,
	}
}

// Put puts the result of the check to the manager in the managers buffer
func (t *TableOutputManager) Put(cr CheckResult) error {
	printResults := func(r Result, prefix string, filename string) {
		d := []string{prefix, filename, r.Error()}
		t.table.Append(d)
		for _, trace := range r.Traces {
			dt := []string{"trace", filename, trace.Error()}
			t.table.Append(dt)
		}
	}

	for _, r := range cr.Successes {
		printResults(r, "success", cr.FileName)
	}

	for _, r := range cr.Warnings {
		printResults(r, "warning", cr.FileName)
	}

	for _, r := range cr.Failures {
		printResults(r, "failure", cr.FileName)
	}

	return nil
}

// Flush writes the contents of the managers buffer to the console
func (t *TableOutputManager) Flush() error {
	if t.table.NumLines() > 0 {
		t.table.Render()
	}

	return nil
}
