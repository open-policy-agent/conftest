package push

import "testing"

func Test_GetRepositoryWithTag_WithoutTag_TagsRepository(t *testing.T) {
	input := "repository"

	expected := "repository:latest"
	actual := getRepositoryWithTag(input)

	if expected != actual {
		t.Error("repository was not tagged. expected %s actual %s", expected, actual)
	}
}

func Test_GetRepositoryWithTag_WithTag_ReturnsSameRepository(t *testing.T) {
	input := "repository:latest"

	expected := "repository:latest"
	actual := getRepositoryWithTag(input)

	if expected != actual {
		t.Error("repository was not tagged. expected %s actual %s", expected, actual)
	}
}
