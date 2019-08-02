package report

import "fmt"

// Reporter controls how results are reported
type Reporter interface {
	Report(results <-chan Result) error
}

// Result represents a test result to be reported
type Result struct {
	Level    Level `json:"level"`
	FileName string `json:"filename"`
	Msg      string `json:"msg"`
}

// Level represents output level (e.g. warn or error)
type Level int

const (
	// Warn level
	Warn Level = iota
	// Error level
	Error
)

const (
	outputSTD  = "stdout"
	outputJSON = "json"
)

func ValidOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
	}
}

func GetReporter(outFmt string, color bool) Reporter {
	switch outFmt {
	case outputSTD:
		return NewDefaultStdOutReporter(color)
	case outputJSON:
		return NewDefaultJSONReporter()
	default:
		return NewDefaultStdOutReporter(color)
	}
}

func getIndicatorForFile(fileName string) string {
	if fileName == "-" {
		return " - "
	}

	return fmt.Sprintf(" - %s - ", fileName)
}
