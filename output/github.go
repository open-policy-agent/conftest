package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/v1/tester"
)

type githubLevel string

const (
	githubInfo  githubLevel = "notice"
	githubWarn  githubLevel = "warning"
	githubError githubLevel = "error"
)

// GitHub represents an Outputter that outputs
// results in GitHub workflow format.
// https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions
type GitHub struct {
	writer io.Writer
}

// NewGitHub creates a new GitHub with the given writer.
func NewGitHub(w io.Writer) *GitHub {
	github := GitHub{
		writer: w,
	}

	return &github
}

// Output outputs the results.
func (g *GitHub) Output(checkResults CheckResults) error {
	var totalFailures int
	var totalExceptions int
	var totalWarnings int
	var totalSuccesses int
	var totalSkipped int
	for _, result := range checkResults {
		totalFailures += len(result.Failures)
		totalExceptions += len(result.Exceptions)
		totalWarnings += len(result.Warnings)
		totalSkipped += len(result.Skipped)
		totalSuccesses += result.Successes

		numPolicies := result.Successes + len(result.Failures) + len(result.Warnings) + len(result.Exceptions) + len(result.Skipped)

		fileLoc := &Location{File: result.FileName, Line: json.Number("1")}

		g.writeLn("::group::Testing %q against %d policies in namespace %q", result.FileName, numPolicies, result.Namespace)
		for _, failure := range result.Failures {
			g.writeLocs(githubError, fileLoc, failure.Location, failure.Message)
		}
		for _, warning := range result.Warnings {
			g.writeLocs(githubWarn, fileLoc, warning.Location, warning.Message)
		}
		for _, exception := range result.Exceptions {
			g.writeLocs(githubInfo, fileLoc, exception.Location, exception.Message)
		}
		for _, skipped := range result.Skipped {
			g.writeLocs(githubInfo, fileLoc, skipped.Location, "Test was skipped: %s", skipped.Message)
		}
		g.writeLoc(githubInfo, fileLoc, "Number of successful checks: %d", result.Successes)
		g.writeLn("::endgroup::")
	}

	totalTests := totalFailures + totalExceptions + totalWarnings + totalSuccesses + totalSkipped

	g.writeLn("%d %s, %d passed, %d %s, %d %s, %d %s",
		totalTests, plural("test", totalTests),
		totalSuccesses,
		totalWarnings, plural("warning", totalWarnings),
		totalFailures, plural("failure", totalFailures),
		totalExceptions, plural("exception", totalExceptions),
	)

	return nil
}

func (g *GitHub) writeLn(msg string, args ...any) {
	fmt.Fprintf(g.writer, msg+"\n", args...)
}

func (g *GitHub) writeLoc(level githubLevel, loc *Location, msg string, args ...any) {
	msg = fmt.Sprintf("::%s file=%s,line=%s::%s", level, loc.File, loc.Line, msg)
	g.writeLn(msg, args...)
}

func (g *GitHub) writeLocs(level githubLevel, fileLoc, ogLoc *Location, msg string, args ...any) {
	// If no location was specified by the policy, default to the file location.
	if ogLoc == nil {
		g.writeLoc(level, fileLoc, msg, args...)
		return
	}

	// If in the same file, prefer the location produced by the Rego policy.
	if ogLoc.File == fileLoc.File {
		g.writeLoc(level, ogLoc, msg, args...)
		return
	}

	// If different files, produce messages for both locations.
	// Always produce a relattive path as some inputs may be a long absolute path.
	og := &Location{
		File: relPath(ogLoc.File),
		Line: ogLoc.Line,
	}
	g.writeLoc(level, og, msg, args...)
	g.writeLoc(level, fileLoc, fmt.Sprintf("(ORIGINATING FROM %s) %s", og, msg), args...)
}

func (g *GitHub) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in GitHub output")
}
