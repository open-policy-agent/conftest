package report

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
)

// JSONReporter reports messages to stdout as a json array
type JSONReporter struct {
	logger *log.Logger

	data []Result
}

func NewDefaultJSONReporter() *JSONReporter {
	return NewJSONReporter(log.New(os.Stdout, "", 0))
}

func NewJSONReporter(l *log.Logger) *JSONReporter {
	return &JSONReporter{
		logger: l,
	}
}

// Report messages in the following format
// { "level": }
func (r *JSONReporter) Report(results <-chan Result) error {
	for result := range results {
		r.data = append(r.data, result)
	}

	b, err := json.Marshal(r.data)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	r.logger.Print(out.String())
	return nil
}
