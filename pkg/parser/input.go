package parser

import (
	"path/filepath"
	"strings"
)

// ValidInputs returns string array in order to passing valid input types to viper
func ValidInputs() []string {
	return []string{
		"toml",
		"tf|hcl",
		"cue",
		"ini",
		"yaml",
	}
}

// Input is the struct that used for deciding which type of parser will be applied
type Input struct {
	input string
	fName string
}

// GetInput returns a valid input object for given fileName and inputType
func GetInput(fileName string, inputType string) *Input {
	if inputType != "" {
		return &Input{
			input: inputType,
			fName: fileName,
		}
	}
	suffix := strings.Replace(filepath.Ext(fileName), ".", "", -1)
	return &Input{
		input: suffix,
		fName: fileName,
	}
}
