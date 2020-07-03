package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jstemmer/go-junit-report/parser"
	"github.com/logrusorgru/aurora"
	table "github.com/olekukonko/tablewriter"
)

const (
	outputSTD   = "stdout"
	outputJSON  = "json"
	outputTAP   = "tap"
	outputTable = "table"
	outputJUnit = "junit"
)

// ValidOutputs returns the available output formats for reporting tests
func ValidOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
		outputTAP,
		outputTable,
		outputJUnit,
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
	case outputJUnit:
		return NewDefaultJUnitOutputManager()
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
	var totalFailures int
	var totalExceptions int
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

		for _, r := range cr.Exceptions {
			printResults(r, "EXCP", aurora.CyanFg)
		}

		totalFailures += len(cr.Failures)
		totalExceptions += len(cr.Exceptions)
		totalWarnings += len(cr.Warnings)
		totalSuccesses += len(cr.Successes)
	}

	totalPolicies := totalFailures + totalExceptions + totalWarnings + totalSuccesses

	var outputColor aurora.Color
	if totalFailures > 0 {
		outputColor = aurora.RedFg
	} else if totalWarnings > 0 {
		outputColor = aurora.YellowFg
	} else if totalExceptions > 0 {
		outputColor = aurora.CyanFg
	} else {
		outputColor = aurora.GreenFg
	}

	var pluralSuffixTests string
	if totalPolicies != 1 {
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

	s.logger.Println()
	outputText := fmt.Sprintf("%v test%s, %v passed, %v warning%s, %v failure%s, %v exception%s",
		totalPolicies, pluralSuffixTests,
		totalSuccesses,
		totalWarnings, pluralSuffixWarnings,
		totalFailures, pluralSuffixFailures,
		totalExceptions, pluralSuffixExceptions,
	)
	s.logger.Println(s.color.Colorize(outputText, outputColor))

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
		result.Warnings = append(result.Warnings, jsonResult{
			Message:  warning.Message,
			Metadata: warning.Metadata,
			Traces:   errsToStrings(warning.Traces),
		})
	}

	for _, failure := range cr.Failures {
		result.Failures = append(result.Failures, jsonResult{
			Message:  failure.Message,
			Metadata: failure.Metadata,
			Traces:   errsToStrings(failure.Traces),
		})
	}

	for _, successes := range cr.Successes {
		result.Successes = append(result.Successes, jsonResult{
			Message:  successes.Message,
			Metadata: successes.Metadata,
			Traces:   errsToStrings(successes.Traces),
		})
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

// JUnitOutputManager formats its output as a JUnit test result
type JUnitOutputManager struct {
	p      parser.Package
	writer io.Writer
}

// NewDefaultJUnitOutputManager creates a new JUnitOutputManager using standard out
func NewDefaultJUnitOutputManager() *JUnitOutputManager {
	return NewJUnitOutputManager(os.Stdout)
}

// NewJUnitOutputManager creates a new JUnitOutputManager with a given Writer
func NewJUnitOutputManager(w io.Writer) *JUnitOutputManager {
	return &JUnitOutputManager{
		writer: w,
		p: parser.Package{
			Name:  "conftest",
			Tests: []*parser.Test{},
		},
	}
}

// Put puts the result of the check to the manager in the managers buffer
func (j *JUnitOutputManager) Put(cr CheckResult) error {
	getOutput := func(r Result) []string {
		out := []string{
			r.Message,
		}
		for _, err := range r.Traces {
			out = append(out, err.Error())
		}
		return out
	}
	convert := func(r Result, status parser.Result) *parser.Test {
		// We have to make sure that the name of the test is unique
		name := fmt.Sprintf(
			"%s - %s",
			cr.FileName,
			strings.Split(r.Message, "\n")[0],
		)

		return &parser.Test{
			Name:   name,
			Result: status,
			Output: getOutput(r),
		}
	}

	for _, result := range cr.Warnings {
		j.p.Tests = append(j.p.Tests, convert(result, parser.FAIL))
	}
	for _, result := range cr.Failures {
		j.p.Tests = append(j.p.Tests, convert(result, parser.FAIL))
	}
	for _, result := range cr.Successes {
		j.p.Tests = append(j.p.Tests, convert(result, parser.PASS))
	}
	return nil
}

// Flush writes the contents of the managers buffer to the console
func (j *JUnitOutputManager) Flush() error {
	report := &parser.Report{
		Packages: []parser.Package{
			j.p,
		},
	}
	return formatter.JUnitReportXML(report, false, runtime.Version(), j.writer)
}
