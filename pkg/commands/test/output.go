package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/logrusorgru/aurora"
)

const (
	outputSTD  = "stdout"
	outputJSON = "json"
	outputTAP  = "tap"
)

func validOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
		outputTAP,
	}
}

func getOutputManager(outFmt string, color bool) outputManager {
	switch outFmt {
	case outputSTD:
		return newDefaultStdOutputManager(color)
	case outputJSON:
		return newDefaultJSONOutputManager()
	case outputTAP:
		return newDefaultTAPOutputManager()
	default:
		return newDefaultStdOutputManager(color)
	}
}

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

// jsonOutputManager reports `conftest` results to `stdout` as a json array..
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
	// we explicitly use an empty slice here to ensure that this field will not be
	// null in json
	res := []string{}
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

// tapOutputManager reports `conftest` results to stdout.
type tapOutputManager struct {
	logger *log.Logger
}

// newDefaultTapOutManager instantiates a new instance of tapOutputManager
// using the default logger.
func newDefaultTAPOutputManager() *tapOutputManager {
	return newTAPOutputManager(log.New(os.Stdout, "", 0))
}

// newTapOutputManager constructs an instance of stdOutputManager given a
// logger instance.
func newTAPOutputManager(l *log.Logger) *tapOutputManager {
	return &tapOutputManager{
		logger: l,
	}
}

func (s *tapOutputManager) put(fileName string, cr checkResult) error {
	var indicator string
	if fileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", fileName)
	}

	issues := len(cr.failures) + len(cr.warnings)
	if issues > 0 {
		s.logger.Print(fmt.Sprintf("1..%d", issues))
		for i, r := range cr.failures {
			s.logger.Print("not ok ", i+1, indicator, r)
		}
		if len(cr.warnings) > 0 {
			s.logger.Print("# Warnings")
			for i, r := range cr.warnings {
				counter := i + 1 + len(cr.failures)
				s.logger.Print("not ok ", counter, indicator, r)
			}
		}
	}

	return nil
}

func (s *tapOutputManager) flush() error {
	// no op
	return nil
}
