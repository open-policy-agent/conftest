package kubernetes

import "encoding/yaml"

command test: {
	task conftest: {
		kind:   "exec"
		cmd:    "conftest test -"
		stdin:  yaml.MarshalStream(objects)
	}
}
