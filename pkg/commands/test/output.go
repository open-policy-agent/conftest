package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/logrusorgru/aurora"
)

// outputManager controls how results of the `ccheck` evaluation will be recorded
// and reported to the end user.
type outputManager interface {
	put(fileName string, cr checkResult) error
	flush() error
}

// stdOutputManager reports `ccheck` results to stdout.
type stdOutputManager struct {
	logger *log.Logger
	color  aurora.Aurora
}

// newDefaultStdOutputManager instantiates a new instance of stdOutputManager
// using the default logger.
func newDefaultStdOutputManager(color bool) *stdOutputManager {
	return newStdOutputManager(log.New(os.Stdout, "", 0), color)
}

// newStdOutputManager constructs an instance of stdOutputManager given a
// logger instance.
func newStdOutputManager(l *log.Logger, color bool) *stdOutputManager {
	return &stdOutputManager{
		logger: l,
		// control color output within the logger
		color: aurora.NewAurora(color),
	}
}

func (s *stdOutputManager) put(fileName string, cr checkResult) error {
	var indicator string
	if fileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", fileName)
	}

	// print warnings and then print errors
	for _, r := range cr.warnings {
		s.logger.Print(s.color.Colorize("WARN", aurora.YellowFg), indicator, r)
	}

	for _, r := range cr.failures {
		s.logger.Print(s.color.Colorize("FAIL", aurora.RedFg), indicator, r)
	}

	return nil
}

func (s *stdOutputManager) flush() error {
	// no op
	return nil
}

type jsonCheckResult struct {
	Filename string   `json:"filename"`
	Warnings []string `json:"warnings"`
	Failures []string `json:"failures"`
}

// jsonOutputManager reports `ccheck` results to `stdout` as a json array..
type jsonOutputManager struct {
	logger *log.Logger

	data []jsonCheckResult
}

func newDefaultJSONOutputManager() *jsonOutputManager {
	return newJSONOutputManager(log.New(os.Stdout, "", 0))
}

func newJSONOutputManager(l *log.Logger) *jsonOutputManager {
	return &jsonOutputManager{
		logger: l,
	}
}

func errsToStrings(errs []error) []string {
	var res []string
	for _, err := range errs {
		res = append(res, err.Error())
	}

	return res
}

func (j *jsonOutputManager) put(fileName string, cr checkResult) error {

	if fileName == "-" {
		fileName = ""
	}

	j.data = append(j.data, jsonCheckResult{
		Filename: fileName,
		Warnings: errsToStrings(cr.warnings),
		Failures: errsToStrings(cr.failures),
	})

	return nil
}

func (j *jsonOutputManager) flush() error {
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
