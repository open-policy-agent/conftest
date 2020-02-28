package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/logrusorgru/aurora"
	table "github.com/olekukonko/tablewriter"
)

const (
	outputSTD   = "stdout"
	outputJSON  = "json"
	outputTAP   = "tap"
	outputTable = "table"
)

// ValidOutputs returns the available output formats for reporting tests
func ValidOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
		outputTAP,
		outputTable,
	}
}

// GetOutputManager returns the OutputManager based on the user input
func GetOutputManager(outputFormat string, color bool) OutputManager {
	switch outputFormat {
	case outputSTD:
		return NewDefaultStandardOutputManager(color)
	case outputJSON:
		return NewDefaultJSONOutputManager()
	case outputTAP:
		return NewDefaultTAPOutputManager()
	case outputTable:
		return NewDefaultTableOutputManager()
	default:
		return NewDefaultStandardOutputManager(color)
	}
}

// OutputManager controls how results of an evaluation will be recorded and reported to the end user
type OutputManager interface {
	Put(cr CheckResult) error
	Flush() error
}

// StandardOutputManager writes to stdout
type StandardOutputManager struct {
	logger  *log.Logger
	color   aurora.Aurora
	results []CheckResult
}

// NewDefaultStandardOutputManager creates a new StandardOutputManager using the default logger
func NewDefaultStandardOutputManager(color bool) *StandardOutputManager {
	return NewStandardOutputManager(log.New(os.Stdout, "", 0), color)
}

// NewStandardOutputManager creates a new StandardOutputManager given a logger instance
func NewStandardOutputManager(l *log.Logger, color bool) *StandardOutputManager {
	return &StandardOutputManager{
		logger: l,
		color:  aurora.NewAurora(color),
	}
}

// Put puts the result of the check to the manager in the managers buffer
func (s *StandardOutputManager) Put(cr CheckResult) error {
	s.results = append(s.results, cr)
	return nil
}

// Flush writes the contents of the managers buffer to the console
func (s *StandardOutputManager) Flush() error {
	var totalPolicies int
	var totalFailures int
	var totalWarnings int
	var totalSuccesses int

	for _, cr := range s.results {
		var indicator string
		if cr.FileName == "-" {
			indicator = " - "
		} else {
			indicator = fmt.Sprintf(" - %s - ", cr.FileName)
		}

		currentPolicies := len(cr.Successes) + len(cr.Warnings) + len(cr.Failures)
		if currentPolicies == 0 {
			s.logger.Print(s.color.Colorize("?", aurora.WhiteFg), indicator, "no policies found")
			continue
		}

		printResults := func(r Result, prefix string, color aurora.Color) {
			s.logger.Print(s.color.Colorize(prefix, color), indicator, r.Message)
			for _, t := range r.Traces {
				s.logger.Print(s.color.Colorize("TRAC", aurora.BlueFg), indicator, t)
			}
		}

		for _, r := range cr.Successes {
			if len(r.Traces) == 0 {
				continue
			}

			printResults(r, "PASS", aurora.GreenFg)
		}

		for _, r := range cr.Warnings {
			printResults(r, "WARN", aurora.YellowFg)
		}

		for _, r := range cr.Failures {
			printResults(r, "FAIL", aurora.RedFg)
		}

		totalPolicies += currentPolicies
		totalFailures += len(cr.Failures)
		totalWarnings += len(cr.Warnings)
		totalSuccesses += len(cr.Successes)
	}

	s.logger.Print("--------------------------------------------------------------------------------")
	s.logger.Print("PASS: ", totalSuccesses, "/", totalPolicies)
	s.logger.Print("WARN: ", totalWarnings, "/", totalPolicies)
	s.logger.Print("FAIL: ", totalFailures, "/", totalPolicies)

	return nil
}

type jsonResult struct {
	Message  string                 `json:"msg"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Traces   []string               `json:"traces,omitempty"`
}

type jsonCheckResult struct {
	Filename  string       `json:"filename"`
	Warnings  []jsonResult `json:"warnings"`
	Failures  []jsonResult `json:"failures"`
	Successes []jsonResult `json:"successes"`
}

// JSONOutputManager formats its output to JSON
type JSONOutputManager struct {
	logger *log.Logger
	data   []jsonCheckResult
}

// NewDefaultJSONOutputManager creates a new JSONOutputManager using the default logger
func NewDefaultJSONOutputManager() *JSONOutputManager {
	return NewJSONOutputManager(log.New(os.Stdout, "", 0))
}

// NewJSONOutputManager creates a new JSONOutputManager with a given logger instance
func NewJSONOutputManager(l *log.Logger) *JSONOutputManager {
	return &JSONOutputManager{
		logger: l,
	}
}

func errsToStrings(errs []error) []string {
	res := []string{}
	for _, err := range errs {
		res = append(res, err.Error())
	}

	return res
}

// Put puts the result of the check to the manager in the managers buffer
func (j *JSONOutputManager) Put(cr CheckResult) error {
	if cr.FileName == "-" {
		cr.FileName = ""
	}

	result := jsonCheckResult{
		Filename:  cr.FileName,
		Warnings:  []jsonResult{},
		Failures:  []jsonResult{},
		Successes: []jsonResult{},
	}

	for _, warning := range cr.Warnings {
		if len(warning.Traces) > 0 {
			result.Warnings = append(result.Warnings, jsonResult{
				Message:  warning.Message,
				Metadata: warning.Metadata,
				Traces:   errsToStrings(warning.Traces),
			})
		} else {
			result.Warnings = append(result.Warnings, jsonResult{
				Message:  warning.Message,
				Metadata: warning.Metadata,
			})
		}
	}

	for _, failure := range cr.Failures {
		if len(failure.Traces) > 0 {
			result.Failures = append(result.Failures, jsonResult{
				Message:  failure.Message,
				Metadata: failure.Metadata,
				Traces:   errsToStrings(failure.Traces),
			})
		} else {
			result.Failures = append(result.Failures, jsonResult{
				Message:  failure.Message,
				Metadata: failure.Metadata,
			})
		}
	}

	for _, successes := range cr.Successes {
		if len(successes.Traces) > 0 {
			result.Successes = append(result.Successes, jsonResult{
				Message:  successes.Message,
				Metadata: successes.Metadata,
				Traces:   errsToStrings(successes.Traces),
			})
		} else {
			result.Successes = append(result.Successes, jsonResult{
				Message:  successes.Message,
				Metadata: successes.Metadata,
			})
		}
	}

	j.data = append(j.data, result)

	return nil
}

// Flush writes the contents of the managers buffer to the console
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

// TAPOutputManager formats its output in TAP format
type TAPOutputManager struct {
	logger *log.Logger
}

// NewDefaultTAPOutputManager creates a new TAPOutputManager using the default logger
func NewDefaultTAPOutputManager() *TAPOutputManager {
	return NewTAPOutputManager(log.New(os.Stdout, "", 0))
}

// NewTAPOutputManager creates a new TAPOutputManager with a given logger instance
func NewTAPOutputManager(l *log.Logger) *TAPOutputManager {
	return &TAPOutputManager{
		logger: l,
	}
}

// Put puts the result of the check to the manager in the managers buffer
func (t *TAPOutputManager) Put(cr CheckResult) error {
	var indicator string
	if cr.FileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", cr.FileName)
	}

	printResults := func(r Result, prefix string, counter int) {
		t.logger.Print(prefix, counter, indicator, r.Message)
		if len(r.Traces) > 0 {
			t.logger.Print("# Traces")
			for j, trace := range r.Traces {
				t.logger.Print("trace ", counter, j+1, indicator, trace.Error())
			}
		}
	}

	issues := len(cr.Failures) + len(cr.Warnings) + len(cr.Successes)
	if issues > 0 {
		t.logger.Print(fmt.Sprintf("1..%d", issues))
		for i, r := range cr.Failures {
			printResults(r, "not ok ", i+1)

		}
		if len(cr.Warnings) > 0 {
			t.logger.Print("# Warnings")
			for i, r := range cr.Warnings {
				counter := i + 1 + len(cr.Failures)
				printResults(r, "not ok ", counter)
			}
		}
		if len(cr.Successes) > 0 {
			t.logger.Print("# Successes")
			for i, r := range cr.Successes {
				counter := i + 1 + len(cr.Failures) + len(cr.Warnings)
				printResults(r, "ok ", counter)
			}
		}
	}

	return nil
}

// Flush is currently a NOOP
func (t *TAPOutputManager) Flush() error {
	return nil
}

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
