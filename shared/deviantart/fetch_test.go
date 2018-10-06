package deviantart

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractDownloadLinkURL(t *testing.T) {
	test := func(link, expected string) {
		// SETUP SUT
		reader := strings.NewReader(link)
		tokenizer := html.NewTokenizer(reader)
		tokenizer.Next()

		// EXERCISE
		actual := extractDownloadLinkURL(tokenizer)

		// VERIFY
		if expected != actual {
			t.Errorf("Expected '%s' but got '%s'", expected, actual)
		}
	}

	t.Run("Class first", func(t *testing.T) {
		test(`<a class="dev-page-download" href="classFirst"/>`, "classFirst")
	})
	t.Run("Href first", func(t *testing.T) {
		test(`<a href="hrefFirst" class=" dev-page-download"/>`, "hrefFirst")
	})
	t.Run("Just a", func(t *testing.T) {
		test("<a/>", "")
	})
	t.Run("Wrong class", func(t *testing.T) {
		test(`<a href="wontBeReturned" class="not-the-right-one"`, "")
	})
}
