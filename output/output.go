package output

const (
	outputSTD   = "stdout"
	outputJSON  = "json"
	outputTAP   = "tap"
	outputTable = "table"
	outputJUnit = "junit"
)

// ValidOutputs returns the available output formats for reporting tests.
func ValidOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
		outputTAP,
		outputTable,
		outputJUnit,
	}
}

// OutputManager controls how results of an evaluation will be recorded and reported to the end user.
type OutputManager interface {
	Put(cr CheckResult) error
	Flush() error
}

// GetOutputManager returns the OutputManager based on the user input.
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
