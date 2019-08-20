package terraform

import "github.com/hashicorp/hcl"

type Parser struct{}

//Format returns the expected format of the input to be parsed
func (s *Parser) Format() string {
	return "terraform"
}

func (s *Parser) Unmarshal(p []byte, v interface{}) error {
	return hcl.Unmarshal(p, v)
}
