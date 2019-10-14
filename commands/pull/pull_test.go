package pull

import (
	"reflect"
	"testing"

	"github.com/instrumenta/conftest/policy"
)

func TestPoliciesToPull(t *testing.T) {

	repositories := []string{
		"my.url.com/repository",
		"my.url.com/repository:v1",
	}

	expected := []policy.Policy{
		{Repository: "my.url.com/repository"},
		{Repository: "my.url.com/repository:v1"},
	}

	actual := getPolicies(repositories)

	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
