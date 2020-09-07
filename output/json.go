package output

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
)

type jsonResult struct {
	Message  string                 `json:"msg"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Traces   []string               `json:"traces,omitempty"`
}

type jsonCheckResult struct {
	Filename  string       `json:"filename"`
	Successes int          `json:"successes"`
	Warnings  []jsonResult `json:"warnings"`
	Failures  []jsonResult `json:"failures"`
}

// JSONOutputManager formats its output to JSON
type JSONOutputManager struct {
	logger *log.Logger
	data   []jsonCheckResult
}

// NewDefaultJSONOutputManager creates a new JSONOutputManager using the default logger
func NewDefaultJSONOutputManager() *JSONOutputManager {
	return NewJSONOutputManager(log.New(os.Stdout, "", 0))
}

// NewJSONOutputManager creates a new JSONOutputManager with a given logger instance
func NewJSONOutputManager(l *log.Logger) *JSONOutputManager {
	return &JSONOutputManager{
		logger: l,
	}
}

// Put puts the result of the check to the manager in the managers buffer
func (j *JSONOutputManager) Put(cr CheckResult) error {
	if cr.FileName == "-" {
		cr.FileName = ""
	}

	result := jsonCheckResult{
		Filename:  cr.FileName,
		Successes: 0,
		Warnings:  []jsonResult{},
		Failures:  []jsonResult{},
	}

	for _, warning := range cr.Warnings {
		result.Warnings = append(result.Warnings, jsonResult{
			Message:  warning.Message,
			Metadata: warning.Metadata,
			Traces:   errsToStrings(warning.Traces),
		})
	}

	for _, failure := range cr.Failures {
		result.Failures = append(result.Failures, jsonResult{
			Message:  failure.Message,
			Metadata: failure.Metadata,
			Traces:   errsToStrings(failure.Traces),
		})
	}

	result.Successes = len(cr.Successes)
	j.data = append(j.data, result)

	return nil
}

// Flush writes the contents of the managers buffer to the console
func (j *JSONOutputManager) Flush() error {
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

func errsToStrings(errs []error) []string {
	res := []string{}
	for _, err := range errs {
		res = append(res, err.Error())
	}

	return res
}
