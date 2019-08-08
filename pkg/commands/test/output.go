package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/viper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
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

func GetOutputManager() OutputManager {
	outFmt := viper.GetString("output")
	color := !viper.GetBool("no-color")

	switch outFmt {
	case outputSTD:
		return newDefaultStdOutputManager(color)
	case outputJSON:
		return newDefaultJSONOutputManager()
	default:
		return newDefaultStdOutputManager(color)
	}
}

// OutputManager controls how results of the `ccheck` evaluation will be recorded
// and reported to the end user.
//counterfeiter:generate . OutputManager
type OutputManager interface {
	Put(fileName string, cr CheckResult) error
	Flush() error
}

// stdOutputManager reports `ccheck` results to stdout.
type stdOutputManager struct {
	logger *log.Logger
	color  aurora.Aurora
}

// newDefaultStdOutputManager instantiates a new instance of stdOutputManager
// using the default logger.
func newDefaultStdOutputManager(color bool) *stdOutputManager {
	return NewStdOutputManager(log.New(os.Stdout, "", 0), color)
}

// NewStdOutputManager constructs an instance of stdOutputManager given a
// logger instance.
func NewStdOutputManager(l *log.Logger, color bool) *stdOutputManager {
	return &stdOutputManager{
		logger: l,
		// control color output within the logger
		color: aurora.NewAurora(color),
	}
}

func (s *stdOutputManager) Put(fileName string, cr CheckResult) error {
	var indicator string
	if fileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", fileName)
	}

	// print warnings and then print errors
	for _, r := range cr.Warnings {
		s.logger.Print(s.color.Colorize("WARN", aurora.YellowFg), indicator, r)
	}

	for _, r := range cr.Failures {
		s.logger.Print(s.color.Colorize("FAIL", aurora.RedFg), indicator, r)
	}

	return nil
}

func (s *stdOutputManager) Flush() error {
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

func (j *jsonOutputManager) Put(fileName string, cr CheckResult) error {

	if fileName == "-" {
		fileName = ""
	}

	j.data = append(j.data, jsonCheckResult{
		Filename: fileName,
		Warnings: errsToStrings(cr.Warnings),
		Failures: errsToStrings(cr.Failures),
	})

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

func (s *tapOutputManager) put(fileName string, cr CheckResult) error {
	var indicator string
	if fileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", fileName)
	}

	issues := len(cr.Failures) + len(cr.Warnings)
	if issues > 0 {
		s.logger.Print(fmt.Sprintf("1..%d", issues))
		for i, r := range cr.Failures {
			s.logger.Print("not ok ", i+1, indicator, r)
		}
		if len(cr.Warnings) > 0 {
			s.logger.Print("# Warnings")
			for i, r := range cr.Warnings {
				counter := i + 1 + len(cr.Failures)
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
