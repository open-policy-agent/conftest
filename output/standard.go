package output

import (
	"fmt"
	"io"

	"github.com/logrusorgru/aurora"
)

// Standard represents an Outputter that outputs
// results in a human readable format.
type Standard struct {
	Writer io.Writer

	// Tracing will render the trace results of the
	// queries when set to true.
	Tracing bool

	// NoColor will disable all coloring when
	// set to true.
	NoColor bool
}

// NewStandard creates a new Standard with the given writer.
func NewStandard(w io.Writer) *Standard {
	standard := Standard{
		Writer: w,
	}

	return &standard
}

// Output outputs the results.
func (s *Standard) Output(results []CheckResult) error {
	colorizer := aurora.NewAurora(true)
	if s.NoColor {
		colorizer = aurora.NewAurora(false)
	}

	if s.Tracing {
		s.outputTrace(results, colorizer)
		return nil
	}

	var totalFailures int
	var totalExceptions int
	var totalWarnings int
	var totalSuccesses int
	for _, result := range results {
		var indicator string
		if result.FileName == "-" {
			indicator = "-"
		} else {
			indicator = fmt.Sprintf("- %s -", result.FileName)
		}

		totalPolicies := result.Successes + len(result.Warnings) + len(result.Failures) + len(result.Exceptions)
		if totalPolicies == 0 {
			fmt.Fprintln(s.Writer, colorizer.Colorize("?", aurora.WhiteFg), indicator, "no policies found")
			continue
		}

		for _, warning := range result.Warnings {
			fmt.Fprintln(s.Writer, colorizer.Colorize("WARN", aurora.YellowFg), indicator, warning.Message)
		}

		for _, failure := range result.Failures {
			fmt.Fprintln(s.Writer, colorizer.Colorize("FAIL", aurora.RedFg), indicator, failure.Message)
		}

		for _, exception := range result.Exceptions {
			fmt.Fprintln(s.Writer, colorizer.Colorize("EXCP", aurora.CyanFg), indicator, exception.Message)
		}

		totalFailures += len(result.Failures)
		totalExceptions += len(result.Exceptions)
		totalWarnings += len(result.Warnings)
		totalSuccesses += result.Successes
	}

	totalTests := totalFailures + totalExceptions + totalWarnings + totalSuccesses

	var pluralSuffixTests string
	if totalTests != 1 {
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

	outputText := fmt.Sprintf("%v test%s, %v passed, %v warning%s, %v failure%s, %v exception%s",
		totalTests, pluralSuffixTests,
		totalSuccesses,
		totalWarnings, pluralSuffixWarnings,
		totalFailures, pluralSuffixFailures,
		totalExceptions, pluralSuffixExceptions,
	)

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

	fmt.Fprintln(s.Writer)
	fmt.Fprintln(s.Writer, colorizer.Colorize(outputText, outputColor))
	return nil
}

func (s *Standard) outputTrace(results []CheckResult, colorizer aurora.Aurora) {
	for _, result := range results {
		for _, query := range result.Queries {
			var color aurora.Color
			if query.Passed() {
				color = aurora.GreenFg
			} else {
				color = aurora.RedFg
			}

			fmt.Fprintln(s.Writer, colorizer.Colorize("file: "+result.FileName+" | query: "+query.Query, color))

			for _, t := range query.Traces {
				fmt.Fprintln(s.Writer, colorizer.Colorize("TRAC ", aurora.BlueFg), "", t)
			}
		}
	}
}
