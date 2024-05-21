package main

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/denarced/dafavorites/shared/deviantart"
)

func TestDeriveFilename(t *testing.T) {
	run := func(name, prefix, url, expected string) {
		t.Run(name, func(t *testing.T) {
			testza.AssertEqual(t, expected, deriveFilename(prefix, url))
		})
	}

	run("No prefix", "", "http://a.com/me.jpg", "me.jpg")
	run(
		"With URL parameters",
		"dox",
		"http://a.com/me.jpg?param=value", "dox_me.jpg")
	run(
		"With dirs in URL",
		"longer",
		"http://b.com/shut/down/more.jpg",
		"longer_more.jpg")
}

func TestExtractDimensions(t *testing.T) {
	run := func(name, filepath string, expected deviantart.Dimensions) {
		t.Run(name, func(t *testing.T) {
			testza.AssertEqual(t, expected, extractDimensions(filepath))
		})
	}

	run(
		"jpg",
		"testdata/test.jpg",
		deviantart.Dimensions{
			Width:  450,
			Height: 29,
		})
	run(
		"png",
		"testdata/test.png",
		deviantart.Dimensions{
			Width:  431,
			Height: 39,
		})
	run(
		"Nonexistent file",
		"nonexistent.file",
		deviantart.Dimensions{})
	run(
		"Non-image file",
		"dafavorites_test.go",
		deviantart.Dimensions{})
}
