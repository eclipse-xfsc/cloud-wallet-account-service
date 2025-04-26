package common

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("test_var", "test")
	actual := GetEnv("test_var")
	expected := "test"
	if actual != expected {
		t.Errorf("actual %q, expected %q", actual, expected)
	}
	actual = GetEnv("test_var_does_not_exist", "test")
	expected = "test"
	if actual != expected {
		t.Errorf("actual %q, expected %q", actual, expected)
	}
}
