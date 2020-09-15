package output

import (
	"fmt"
	"log"
	"os"

	"github.com/logrusorgru/aurora"
)

// StandardOutputManager writes to stdout
type StandardOutputManager struct {
	logger  *log.Logger
	color   aurora.Aurora
	tracing bool
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

// WithTracing adds tracing to the output.
func (s *StandardOutputManager) WithTracing() OutputManager {
	s.tracing = true
	return s
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
		if cr.Filename == "-" {
			indicator = " - "
		} else {
			indicator = fmt.Sprintf(" - %s - ", cr.Filename)
		}

		currentPolicies := cr.Successes + len(cr.Warnings) + len(cr.Failures) + len(cr.Exceptions)
		if currentPolicies == 0 {
			s.logger.Print(s.color.Colorize("?", aurora.WhiteFg), indicator, "no policies found")
			continue
		}

		printResults := func(r Result, prefix string, color aurora.Color) {
			s.logger.Print(s.color.Colorize(prefix, color), indicator, r.Message)
			if s.tracing {
				for _, t := range r.Traces {
					s.logger.Print(s.color.Colorize("TRAC", aurora.BlueFg), indicator, t)
				}
			}

		}

		for i := 0; i < cr.Successes; i++ {
			if s.tracing {
				printResults(Result{}, "PASS", aurora.GreenFg)
			}
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
		totalSuccesses += cr.Successes
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
