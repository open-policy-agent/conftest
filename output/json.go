package output

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
)

// JSONOutputManager formats its output to JSON.
type JSONOutputManager struct {
	logger  *log.Logger
	data    []CheckResult
	tracing bool
}

// NewDefaultJSONOutputManager creates a new JSONOutputManager using the default logger.
func NewDefaultJSONOutputManager() *JSONOutputManager {
	return NewJSONOutputManager(log.New(os.Stdout, "", 0))
}

// NewJSONOutputManager creates a new JSONOutputManager with a given logger instance.
func NewJSONOutputManager(l *log.Logger) *JSONOutputManager {
	return &JSONOutputManager{
		logger: l,
	}
}

// WithTracing adds tracing to the output.
func (j *JSONOutputManager) WithTracing() OutputManager {
	j.tracing = true
	return j
}

// Put puts the result of the check to the manager in the managers buffer.
func (j *JSONOutputManager) Put(cr CheckResult) error {
	if !j.tracing {
		cr.Queries = nil
	}

	if cr.FileName == "-" {
		cr.FileName = ""
	}

	j.data = append(j.data, cr)
	return nil
}

// Flush writes the contents of the managers buffer to the console.
func (j *JSONOutputManager) Flush() error {
	b, err := json.Marshal(j.data)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	j.logger.Print(out.String())
	return nil
}
