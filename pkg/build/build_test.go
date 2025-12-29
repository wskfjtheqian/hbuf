package build_test

import (
	"hbuf/pkg/build"
	"testing"
)

func Test_StringToMiddleLine(t *testing.T) {
	t.Run("HelloWorld", func(t *testing.T) {
		input := "HelloWorld"
		expected := "hello-world"
		actual := build.StringToMiddleLine(input)
		if actual != expected {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	})
	t.Run("Hello_World", func(t *testing.T) {
		input := "hello_world"
		expected := "hello-world"
		actual := build.StringToMiddleLine(input)
		if actual != expected {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	})
}
