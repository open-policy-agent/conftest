package report

// Reporter controls how results are reported
type Reporter interface {
	Report(level Level, fileName string, msg string)
}

// Level represents output level (e.g. warn or error)
type Level int 

const (
	// Warn level
	Warn Level = iota
	// Error level
	Error
)

func GetReporter(color bool) Reporter {
	return NewDefaultStdOutReporter(color)
}