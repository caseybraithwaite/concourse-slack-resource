package main

import (
	"os"
	"testing"
)

func TestInterpolate(t *testing.T) {
	str := "Hello from <{{ $ATC_EXTERNAL_URL }}|Concourse>"

	os.Setenv("ATC_EXTERNAL_URL", "https://example.org")
	defer os.Unsetenv("ATC_EXTERNAL_URL")

	actual := interpolate(str)
	expected := "Hello from <https://example.org|Concourse>"

	if actual != expected {
		t.Errorf("got %s, wanted %s", actual, expected)
	}
}
