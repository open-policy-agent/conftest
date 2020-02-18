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
	"github.com/spf13/viper"
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

func GetOutputManager() OutputManager {
	outFmt := viper.GetString("output")
	color := !viper.GetBool("no-color")

	switch outFmt {
	case outputSTD:
		return NewDefaultStdOutputManager(color)
	case outputJSON:
		return NewDefaultJSONOutputManager()
	case outputTAP:
		return NewDefaultTAPOutputManager()
	case outputTable:
		return NewDefaultTableOutputManager()
	default:
		return NewDefaultStdOutputManager(color)
	}
}

// OutputManager controls how results of the `ccheck` evaluation will be recorded
// and reported to the end user.
type OutputManager interface {
	Put(cr CheckResult) error
	Flush() error
}

type stdOutputManager struct {
	logger *log.Logger
	color  aurora.Aurora
}

// NewDefaultStdOutputManager instantiates a new instance of stdOutputManager
// using the default logger.
func NewDefaultStdOutputManager(color bool) *stdOutputManager {
	return NewStdOutputManager(log.New(os.Stdout, "", 0), color)
}

// NewStdOutputManager constructs an instance of stdOutputManager given a
// logger instance.
func NewStdOutputManager(l *log.Logger, color bool) *stdOutputManager {
	return &stdOutputManager{
		logger: l,
		color:  aurora.NewAurora(color),
	}
}

func (s *stdOutputManager) Put(cr CheckResult) error {
	var indicator string
	if cr.FileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", cr.FileName)
	}

	printResults := func(r Result, prefix string, color aurora.Color) {
		s.logger.Print(s.color.Colorize(prefix, color), indicator, r.Message)
		for _, t := range r.Traces {
			s.logger.Print(s.color.Colorize("TRAC", aurora.BlueFg), indicator, t)
		}
	}

	// print successes, warnings, errors and their traces
	for _, r := range cr.Successes {
		printResults(r, "PASS", aurora.GreenFg)
	}

	for _, r := range cr.Warnings {
		printResults(r, "WARN", aurora.YellowFg)
	}

	for _, r := range cr.Failures {
		printResults(r, "FAIL", aurora.RedFg)
	}

	return nil
}

func (s *stdOutputManager) Flush() error {
	return nil
}

type jsonResult struct {
	Message map[string]interface{} `json:"message"`
	Traces  []string               `json:"traces,omitempty"`
}

type jsonCheckResult struct {
	Filename  string       `json:"filename"`
	Warnings  []jsonResult `json:"warnings"`
	Failures  []jsonResult `json:"failures"`
	Successes []jsonResult `json:"successes"`
}

// jsonOutputManager reports `conftest` results to `stdout` as a json array..
type jsonOutputManager struct {
	logger *log.Logger

	data []jsonCheckResult
}

func NewDefaultJSONOutputManager() *jsonOutputManager {
	return NewJSONOutputManager(log.New(os.Stdout, "", 0))
}

func NewJSONOutputManager(l *log.Logger) *jsonOutputManager {
	return &jsonOutputManager{
		logger: l,
	}
}

func errsToStrings(errs []error) []string {
	// we explicitly use an empty slice here to ensure that this field will not be
	// null in json
	res := []string{}
	for _, err := range errs {
		res = append(res, err.Error())
	}

	return res
}

func (j *jsonOutputManager) Put(cr CheckResult) error {
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
				Message: warning.Message,
				Traces:  errsToStrings(warning.Traces), // need json result here? create new type thing?? etcetc
			})
		} else {
			result.Warnings = append(result.Warnings, jsonResult{
				Message: warning.Message,
			})
		}
	}

	for _, failure := range cr.Failures {
		if len(failure.Traces) > 0 {
			result.Failures = append(result.Failures, jsonResult{
				Message: failure.Message,
				Traces:  errsToStrings(failure.Traces),
			})
		} else {
			result.Failures = append(result.Failures, jsonResult{
				Message: failure.Message,
			})
		}
	}

	for _, successes := range cr.Successes {
		if len(successes.Traces) > 0 {
			result.Successes = append(result.Successes, jsonResult{
				Message: successes.Message,
				Traces:  errsToStrings(successes.Traces),
			})
		} else {
			result.Successes = append(result.Successes, jsonResult{
				Message: successes.Message,
			})
		}
	}

	j.data = append(j.data, result)

	return nil
}

func (j *jsonOutputManager) Flush() error {
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

// tapOutputManager reports `conftest` results to stdout.
type tapOutputManager struct {
	logger *log.Logger
}

// NewDefaultTAPOutputManager instantiates a new instance of tapOutputManager
// using the default logger.
func NewDefaultTAPOutputManager() *tapOutputManager {
	return NewTAPOutputManager(log.New(os.Stdout, "", 0))
}

// NewTAPOutputManager constructs an instance of stdOutputManager given a
// logger instance.
func NewTAPOutputManager(l *log.Logger) *tapOutputManager {
	return &tapOutputManager{
		logger: l,
	}
}

func (s *tapOutputManager) Put(cr CheckResult) error {

	var indicator string
	if cr.FileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", cr.FileName)
	}

	printResults := func(r Result, prefix string, counter int) {
		s.logger.Print(prefix, counter, indicator, r.Message)
		if len(r.Traces) > 0 {
			s.logger.Print("# Traces")
			for j, t := range r.Traces {
				s.logger.Print("trace ", counter, j+1, indicator, t.Error())
			}
		}
	}

	issues := len(cr.Failures) + len(cr.Warnings) + len(cr.Successes)
	if issues > 0 {
		s.logger.Print(fmt.Sprintf("1..%d", issues))
		for i, r := range cr.Failures {
			printResults(r, "not ok ", i+1)

		}
		if len(cr.Warnings) > 0 {
			s.logger.Print("# Warnings")
			for i, r := range cr.Warnings {
				counter := i + 1 + len(cr.Failures)
				printResults(r, "not ok ", counter)
			}
		}
		if len(cr.Successes) > 0 {
			s.logger.Print("# Successes")
			for i, r := range cr.Successes {
				counter := i + 1 + len(cr.Failures) + len(cr.Warnings)
				printResults(r, "ok ", counter)
			}
		}
	}

	return nil
}

func (s *tapOutputManager) Flush() error {
	return nil
}

type tableOutputManager struct {
	table *table.Table
}

// NewDefaultTableOutputManager instantiates a new instance of tableOutputManager
func NewDefaultTableOutputManager() *tableOutputManager {
	return NewTableOutputManager(os.Stdout)
}

// NewTableOutputManager constructs an instance of tableOutputManager given a
// io.Writer.
func NewTableOutputManager(w io.Writer) *tableOutputManager {
	table := table.NewWriter(w)
	table.SetHeader([]string{"result", "file", "message"})
	return &tableOutputManager{
		table: table,
	}
}

func (s *tableOutputManager) Put(cr CheckResult) error {
	printResults := func(r Result, prefix string, filename string) {
		d := []string{prefix, filename, r.Error()}
		s.table.Append(d)
		for _, t := range r.Traces {
			dt := []string{"trace", filename, t.Error()}
			s.table.Append(dt)
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

func (s *tableOutputManager) Flush() error {
	if s.table.NumLines() > 0 {
		s.table.Render()
	}
	return nil
}
