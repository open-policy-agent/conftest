package kubernetes

import "encoding/yaml"


objects: [ x for x in deployment ]

command dump: {
	task print: {
		kind: "print"
		text: yaml.MarshalStream(objects)
	}
}
