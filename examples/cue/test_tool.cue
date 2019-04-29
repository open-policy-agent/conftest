package kubernetes

import "encoding/yaml"

command test: {
	task conftest: {
		kind:   "exec"
		cmd:    "conftest -"
		stdin:  yaml.MarshalStream(objects)
	}
}
