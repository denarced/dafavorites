package main

import (
	"testing"

	"github.com/denarced/dafavorites/shared/deviantart"
)

func TestDeriveFilename(t *testing.T) {
	var prefix string
	var url string
	var expected string

	t.Run("No prefix", func(t *testing.T) {
		url = "http://a.com/me.jpg"
		expected = "me.jpg"
	})
	t.Run("With URL parameters", func(t *testing.T) {
		prefix = "dox"
		url = "http://a.com/me.jpg?param=value"
		expected = "dox_me.jpg"
	})
	t.Run("With dirs in URL", func(t *testing.T) {
		prefix = "longer"
		url = "http://b.com/shut/down/more.jpg"
		expected = "longer_more.jpg"
	})

	// EXERCISE
	actual := deriveFilename(prefix, url)

	// VERIFY
	if expected != actual {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}

func TestExtractDimensions(t *testing.T) {
	filepath := ""
	expected := deviantart.Dimensions{}

	t.Run("jpg", func(t *testing.T) {
		filepath = "testdata/test.jpg"
		expected = deviantart.Dimensions{
			Width:  450,
			Height: 29,
		}
	})
	t.Run("png", func(t *testing.T) {
		filepath = "testdata/test.png"
		expected = deviantart.Dimensions{
			Width:  431,
			Height: 39,
		}
	})
	t.Run("Nonexistent file", func(t *testing.T) {
		filepath = "nonexistent.file"
		expected = deviantart.Dimensions{}
	})
	t.Run("Non-image file", func(t *testing.T) {
		filepath = "dafavorites_test.go"
		expected = deviantart.Dimensions{}
	})

	// EXERCISE
	actual := extractDimensions(filepath)

	// VERIFY
	if expected != actual {
		t.Errorf("Expected %v dimensions but got %v", expected, actual)
	}
}
