package main

import (
	"testing"

	"github.com/MarvinJWendt/testza"
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
